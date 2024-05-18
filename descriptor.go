package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	maxInputLineLength = 1024
	maxLines           = 50
	maxSubLen          = 128
	MAX_CONNECT        = 15
)

var HTTPGET = []byte("GET ")
var HTTPGETLEN = len(HTTPGET) - 1

var attemptMap map[string]int = make(map[string]int)

// Handle incoming connections.
func handleDesc(conn net.Conn, tls bool) {

	//Parse address
	addr := conn.RemoteAddr().String()
	ipStr, _, _ := net.SplitHostPort(addr)

	//Track connection attempts.
	if attemptMap[ipStr] > MAX_CONNECT || attemptMap[ipStr] == -1 {
		conn.Close()
		return
	} else if attemptMap[ipStr] == MAX_CONNECT {
		conn.Close()
		critLog("Too many connect attempts from %v. Blocking!", ipStr)
		attemptMap[ipStr]++
		return
	}
	attemptMap[ipStr]++

	//Get address of client
	addrList, _ := net.LookupHost(ipStr)
	hostStr := strings.Join(addrList, ", ")

	//If reverse DNS found, make combined string
	cAddr := hostStr
	if hostStr != ipStr {
		cAddr = fmt.Sprintf("%v : %v", ipStr, hostStr)
	}

	//Create descriptor
	descLock.Lock()

	//Incrememnt desc ID, create new descriptor
	topID++
	tnd := telnetData{
		charset: DEFAULT_CHARSET, charMap: DEFAULT_CHARMAP,
		options: &termSettings{},
	}
	desc := &descData{
		conn: conn, id: topID, connectTime: time.Now(),
		reader: bufio.NewReader(conn), tls: tls,
		host: hostStr, addr: ipStr, cAddr: cAddr,
		state: CON_LOGIN, telnet: tnd, valid: true, idleTime: time.Now()}
	descList = append(descList, desc)
	descLock.Unlock()

	//If not TLS, look for HTTP request (TLS fails)
	if !tls {
		conn.SetReadDeadline(time.Now().Add(time.Millisecond))
		data, err := desc.reader.ReadString('\n')
		if err == nil && strings.ContainsAny("GET", data) {
			critLog("HTTP request from %v. Adding to ignore list.", ipStr)
			attemptMap[ipStr] = -1
			conn.Write([]byte(`HTTP/1.1 301 Moved Permanently\r\nLocation: http://www.example.org/`))
			conn.Close()
			return
		}
		conn.SetReadDeadline(time.Time{})
		if len(data) > 0 {
			critLog("Got header: '%v'", string(data))
		}
	}

	critLog("Connection from: %v", ipStr)

	//Send telnet sequences
	sendStart(desc.conn)

	//Launch read loop
	desc.readDescLoop()

	//When loop exits, close
	descLock.Lock()
	desc.close()
	descLock.Unlock()
}

func sendStart(conn net.Conn) {
	//Start telnet negotiation
	sendTelnetCmds(conn)

	//Send greeting
	_, err := conn.Write([]byte(greetBuf))
	if err != nil {
		conn.Close()
		return
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
					errLog("#%v: GOT unknown sub data: %v: %v", desc.id, TermOpt2TXT[int(desc.telnet.subSeqType)], string(desc.telnet.subSeqData))
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

				//Detect line end, ingest line
				if (lastByte == '\r' && data == '\n') ||
					(lastByte != '\r' && data == '\n') {
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
	desc.telnet.termType = filterTelnetResponse(string(desc.telnet.subSeqData))
	match := termTypeMap[desc.telnet.termType]

	//errLog("#%v: GOT %v: %s", desc.id, TermOpt2TXT[int(desc.telnet.subType)], desc.telnet.subData)
	if match != nil {
		desc.telnet.options = match
		if match.CharMap != nil {
			desc.telnet.charMap = match.CharMap
		}
		errLog("Found client match: %v", desc.telnet.termType)
	}
	for n, item := range termTypeMap {
		if strings.HasPrefix(desc.telnet.termType, n) {
			errLog("Found client prefix match: %v", desc.telnet.termType)
			desc.telnet.options = item
			if item.CharMap != nil {
				desc.telnet.charMap = item.CharMap
			}
		} else if strings.HasSuffix(desc.telnet.termType, n) {
			errLog("Found client suffix match: %v", desc.telnet.termType)
			desc.telnet.options = item
			if item.CharMap != nil {
				desc.telnet.charMap = item.CharMap
			}
		}
	}
}

func (desc *descData) getCharset() {
	desc.telnet.charset = filterTelnetResponse(string(desc.telnet.subSeqData))
	desc.setCharset()
	//errLog("#%v: GOT %v: %v", desc.id, TermOpt2TXT[int(desc.telnet.subType)], desc.telnet.charset)

	desc.sendSubSeq(desc.telnet.charset, TermOpt_CHARSET, SB_ACCEPTED)
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
	if !desc.telnet.options.UTF {
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
			desc.telnet.options.SuppressGoAhead = true
		}

	} else if command == TermCmd_DO {

		if option == TermOpt_CHARSET {
			desc.sendSubSeq(charsetSend, TermOpt_CHARSET, SB_REQ)
		} else if option == TermOpt_SUP_GOAHEAD {
			desc.telnet.options.SuppressGoAhead = true
		}

	} else if command == TermCmd_DONT {

		if option == TermOpt_SUP_GOAHEAD {
			desc.telnet.options.SuppressGoAhead = false
		}

	} else if command == TermCmd_WONT {

		if option == TermCmd_GOAHEAD {
			desc.telnet.options.SuppressGoAhead = false
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
