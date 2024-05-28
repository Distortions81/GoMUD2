package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	maxInputLineLength = 1024
	maxLines           = 50
	maxSubLen          = 128
)

func reverseDNS(ip string) string {
	timeout := 10 * time.Second // Timeout duration

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ch := make(chan []string, 1)
	go func() {
		names, err := net.LookupAddr(ip)
		if err != nil {
			errLog("Reverse DNS lookup failed: %v", err)
			ch <- nil
			return
		}
		ch <- names
	}()

	select {
	case names := <-ch:
		for _, name := range names {
			return name
		}
	case <-ctx.Done():
		errLog("Reverse DNS lookup timed out")
	}

	return "unknown"
}

// Handle incoming connections.
func handleDesc(conn net.Conn, tls bool) {

	//Parse address
	a := conn.RemoteAddr().String()
	ip, _, _ := net.SplitHostPort(a)

	//Track connection attempts.
	blockedLock.Lock()
	if blockedMap[ip] != nil {
		blockedMap[ip].Attempts++
		blockedDirty = true
		if blockedMap[ip].Blocked {
			conn.Close()
			blockedLock.Unlock()
			return
		} else if blockedMap[ip].Attempts >= BLOCKED_THRESH {
			conn.Close()
			critLog("Too many connect attempts from %v. Blocking!", ip)
			blockedMap[ip].Blocked = true
			blockedMap[ip].Modified = time.Now().UTC()
			blockedDirty = true
			blockedLock.Unlock()
			return
		}

	} else {
		blockedMap[ip] = &blockedData{Host: ip, Created: time.Now().UTC()}
		blockedDirty = true
	}
	blockedLock.Unlock()

	//Create descriptor
	descLock.Lock()

	//Incrememnt desc ID, create new descriptor
	topID++
	tnd := telnetData{
		Charset: DEFAULT_CHARSET, charMap: DEFAULT_CHARMAP,
		Options: &termSettings{},
	}
	desc := &descData{
		conn: conn, id: topID, connectTime: time.Now(),
		reader: bufio.NewReader(conn), tls: tls, ip: ip,
		state: CON_ACCOUNT, telnet: tnd, valid: true, idleTime: time.Now()}
	descList = append(descList, desc)
	descLock.Unlock()

	go func(desc *descData, ip string) {
		desc.dns = reverseDNS(ip)
	}(desc, ip)

	//If not TLS, look for HTTP request (TLS fails)
	if !tls {
		conn.SetReadDeadline(time.Now().Add(time.Millisecond))
		data, err := desc.reader.ReadString('\n')
		if err == nil && strings.ContainsAny("GET", data) {
			critLog("HTTP request from %v. Adding to ignore list.", ip)
			blockedLock.Lock()
			if blockedMap[ip] == nil {
				blockedMap[ip] = &blockedData{Host: ip, Blocked: true, HTTP: true, Created: time.Now().UTC()}
				blockedDirty = true
			} else {
				blockedMap[ip].Blocked = true
				blockedMap[ip].HTTP = true
				blockedDirty = true
			}
			blockedLock.Unlock()
			conn.Write([]byte(`HTTP/1.1 301 Moved Permanently\r\nLocation: http://www.example.org/`))
			conn.Close()
			return
		}
		conn.SetReadDeadline(time.Time{})
		if len(data) > 0 {
			errLog("Got header: '%v'", string(data))
		}
	}

	mudLog("Connection from: %v", ip)

	//Send telnet sequences
	sendStart(desc.conn, tls)

	//Launch read loop
	desc.readDescLoop()

	//When loop exits, close
	descLock.Lock()
	desc.close()
	descLock.Unlock()
}

func sendStart(conn net.Conn, tls bool) {
	//Start telnet negotiation
	sendTelnetCmds(conn)

	//Send greeting
	var err error
	if tls {
		_, err = conn.Write([]byte(greetBuf))
	} else {
		_, err = conn.Write([]byte(greetBufNoSSL))
	}
	if err != nil {
		conn.Close()
		return
	}
}

func ToAllConnections(format string, args ...any) {
	for _, desc := range descList {
		desc.sendln(format, args...)
	}
}

func (desc *descData) readDescLoop() {

	//Read loop
	var lastByte byte
	for serverState.Load() == SERVER_RUNNING {

		//Read byte
		data, err := desc.readByte()
		if err != nil {
			return
		}

		switch data {

		//Handle telnet escape code
		case TermCmd_IAC:
			//Get telnet command
			command, err := desc.readByte()
			if err != nil {
				return
			}

			//Sub-negotiation sequence END
			if command == TermCmd_SE {

				if desc.telnet.subSeqType == TermOpt_TERMINAL_TYPE {
					desc.getTermType()
				} else if desc.telnet.subSeqType == TermOpt_CHARSET {
					desc.getCharset()
				} else {
					critLog("#%v: GOT unknown sub data: %v: %v", desc.id, TermOpt2TXT[int(desc.telnet.subSeqType)], string(desc.telnet.subSeqData))
				}

				desc.telnet.subSeqMode = false
				desc.telnet.subSeqData = []byte{}
				continue
			}

			//Grab telnet option
			option, err := desc.readByte()
			if err != nil {
				return
			}

			desc.handleTelnetCmd(command, option)

			//errLog("#%v: Client: %v %v", desc.id, TermCmd2Txt[int(command)], TermOpt2TXT[int(option)])
		default:
			if desc.telnet.subSeqMode {
				desc.getSubSeqData(data)
			} else {

				//Limit line length
				if desc.inBufLen > maxInputLineLength {
					desc.inputFull()
					return
				}

				if desc.inBufLen > 0 {
					if data == 127 {
						desc.inBufLen--
						desc.inBuf = desc.inBuf[:desc.inBufLen]
					}
				} else {
					if data == 127 {
						continue
					}
				}

				//Detect line end, ingest line
				if (lastByte == '\r' && data == '\n') ||
					(lastByte == '\n' && data == '\r') ||
					((lastByte != '\r' && lastByte != '\n') && data == '\r') {
					desc.ingestLine()
					continue
				}
				lastByte = data

				//No control chars, no delete, but allow valid UTF-8
				if data >= ' ' && data != 127 {
					desc.inBufLen += 1
					desc.inBuf = append(desc.inBuf, data)
				}
			}
		}
	}
}

func (desc *descData) getTermType() {
	desc.telnet.termType = txtTo7bitUpper(string(desc.telnet.subSeqData))
	match := termTypeMap[desc.telnet.termType]

	//errLog("#%v: GOT %v: %s", desc.id, TermOpt2TXT[int(desc.telnet.subType)], desc.telnet.subData)
	if match != nil {
		desc.telnet.Options = &termSettings{
			ansi256: match.ansi256, ansi24: match.ansi24,
			UTF: match.UTF, suppressGoAhead: match.suppressGoAhead}
		if match.charMap != nil {
			desc.telnet.charMap = match.charMap
		}
		mudLog("Found client match: %v", desc.telnet.termType)
	}
	for n, item := range termTypeMap {
		if strings.HasPrefix(desc.telnet.termType, n) {
			mudLog("Found client prefix match: %v", desc.telnet.termType)
			desc.telnet.Options = &termSettings{
				ansi256: item.ansi256, ansi24: item.ansi24,
				UTF: item.UTF, suppressGoAhead: item.suppressGoAhead}
			if item.charMap != nil {
				desc.telnet.charMap = item.charMap
			}
		} else if strings.HasSuffix(desc.telnet.termType, n) {
			mudLog("Found client suffix match: %v", desc.telnet.termType)
			desc.telnet.Options = &termSettings{
				ansi256: item.ansi256, ansi24: item.ansi24,
				UTF: item.UTF, suppressGoAhead: item.suppressGoAhead}
			if item.charMap != nil {
				desc.telnet.charMap = item.charMap
			}
		}
	}
}

func (desc *descData) getCharset() {
	desc.telnet.Charset = txtTo7bitUpper(string(desc.telnet.subSeqData))
	desc.setCharset()
	//errLog("#%v: GOT %v: %v", desc.id, TermOpt2TXT[int(desc.telnet.subType)], desc.telnet.charset)

	desc.sendSubSeq(desc.telnet.Charset, TermOpt_CHARSET, SB_ACCEPTED)
}

func (desc *descData) getSubSeqData(data byte) {

	if desc.telnet.subLength > maxSubLen {
		desc.telnet.subSeqMode = false
		desc.telnet.subSeqData = []byte{}
		desc.telnet.subLength = 0
		errLog("#%v subsequence size went over %v, abort", desc.id, maxSubLen)
		return
	}

	//7-bit ascii only
	if data >= ' ' && data <= '~' {
		desc.telnet.subSeqData = append(desc.telnet.subSeqData, data)
	}
}

func (desc *descData) ingestLine() {
	desc.inputLock.Lock()
	defer desc.inputLock.Unlock()

	//Too many lines
	if desc.numInputLines > maxLines {
		desc.inputFull()
		return
	}

	//Charmap translation, if needed
	var buf string
	if !desc.telnet.Options.UTF {
		buf = encodeToUTF(desc.telnet.charMap, desc.inBuf)
	} else {
		buf = string(desc.inBuf)
	}

	//Append line to buffer
	desc.inputLines = append(desc.inputLines, buf)
	desc.numInputLines++

	//Reset input buffer
	desc.inBuf = []byte{}
	desc.inBufLen = 0
}

func (desc *descData) handleTelnetCmd(command, option byte) {
	//Sub-negotiation sequence START
	if command == TermCmd_SB {

		desc.telnet.subSeqData = []byte{}
		desc.telnet.subSeqMode = true
		desc.telnet.subSeqType = option

	} else if command == TermCmd_WILL {

		if option == TermOpt_TERMINAL_TYPE {
			desc.sendSubSeq("", TermOpt_TERMINAL_TYPE, SB_SEND)
		} else if option == TermOpt_SUP_GOAHEAD {
			desc.telnet.Options.suppressGoAhead = true
		}

	} else if command == TermCmd_DO {

		if option == TermOpt_CHARSET {
			desc.sendSubSeq(charsetSend, TermOpt_CHARSET, SB_REQ)
		} else if option == TermOpt_SUP_GOAHEAD {
			desc.telnet.Options.suppressGoAhead = true
		}

	} else if command == TermCmd_DONT {

		if option == TermOpt_SUP_GOAHEAD {
			desc.telnet.Options.suppressGoAhead = false
		}

	} else if command == TermCmd_WONT {

		if option == TermCmd_GOAHEAD {
			desc.telnet.Options.suppressGoAhead = false
		}
	}
}

func (desc *descData) send(format string, args ...any) error {
	if desc == nil {
		return nil
	}

	//Format string if args supplied
	var data string
	if args != nil {
		data = fmt.Sprintf(format, args...)
	} else {
		data = format
	}

	desc.outBuf = append(desc.outBuf, data...)
	desc.haveOut = true
	return nil
}

func (desc *descData) sendln(format string, args ...any) error {
	return desc.send(format+"\r\n", args...)
}
