package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	maxInputLineLength = 1024 * 10
	connDeadline       = time.Second * 15
	maxLines           = 500
	maxSubLen          = 128
)

// Handle incoming connections.
func handleDesc(conn net.Conn, tls bool) {
	var tlsStr string

	//Parse address
	rAddr := conn.RemoteAddr().String()
	ipStr, _, _ := net.SplitHostPort(rAddr)
	addrList, _ := net.LookupHost(ipStr)
	hostStr := strings.Join(addrList, ", ")

	//If reverse DNS found, make combined string
	cAddr := hostStr
	if hostStr != ipStr {
		cAddr = fmt.Sprintf("%v : %v", ipStr, hostStr)
	}

	//Create descriptor
	descLock.Lock()
	topID++
	tnd := telnetData{
		charset: DEFAULT_CHARSET, charMap: DEFAULT_CHARMAP,
		options: &termSettings{},
	}
	desc := &descData{
		conn: conn, id: topID, connectTime: time.Now(),
		reader: bufio.NewReader(conn), tls: tls,
		host: hostStr, addr: ipStr, cAddr: cAddr,
		state: CON_WELCOME, telnet: tnd}
	descList = append(descList, desc)
	descLock.Unlock()

	//Close desc on return
	defer desc.close()

	//Connect log message
	if tls {
		tlsStr = " (TLS)"
	}
	mudLog("#%v: %v connected.%v", desc.id, desc.host, tlsStr)

	//Start telnet negotiation
	desc.sendTelnetCmds()

	//Send greeting
	err := desc.sendln(greetBuf)
	desc.state = CON_LOGIN
	if err != nil {
		return
	}

	desc.readDescLoop()
}

func (desc *descData) readDescLoop() {
	//Read loop
	var lastByte byte
	for serverState == SERVER_RUNNING {

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

			//Sub-negotation sequence END
			if command == TermCmd_SE {

				if desc.telnet.subType == TermOpt_TERMINAL_TYPE {
					desc.getTermType()
				} else if desc.telnet.subType == TermOpt_CHARSET {
					desc.getCharset()
				} else {
					errLog("#%v: GOT unknown sub data: %v: %v", desc.id, TermOpt2TXT[int(desc.telnet.subType)], string(desc.telnet.subData))
				}

				desc.telnet.subMode = false
				desc.telnet.subData = []byte{}
				continue
			}

			//Grab telnet option
			option, err := desc.readByte()
			if err != nil {
				return
			}

			desc.handleTelCmd(command, option)

			errLog("#%v: Client: %v %v", desc.id, TermCmd2Txt[int(command)], TermOpt2TXT[int(option)])
		default:
			if desc.telnet.subMode {
				desc.captureSubSeqData(data)
			} else {

				//Limit line length
				if desc.inputBufferLen > maxInputLineLength {
					desc.inputFull()
					return
				}

				//Detect line end
				if (lastByte == '\n' && data == '\r') ||
					(lastByte == '\r' && data == '\n') {
					desc.ingestLine()
					continue
				}

				lastByte = data

				//No control chars, no delete, but allow valid UTF-8
				if data >= ' ' && data != 127 {
					desc.inputBufferLen += 1
					desc.inputBuffer = append(desc.inputBuffer, data)
				}
			}
		}
	}
}

func (desc *descData) getTermType() {
	desc.telnet.termType = telSnFilter(string(desc.telnet.subData))
	match := termTypeMap[desc.telnet.termType]

	errLog("#%v: GOT %v: %s", desc.id, TermOpt2TXT[int(desc.telnet.subType)], desc.telnet.subData)
	if match != nil {
		desc.telnet.options = match
		if match.CharMap != nil {
			desc.telnet.charMap = match.CharMap
		}
		errLog("Found client match: %v", desc.telnet.termType)
	}
	for n, item := range termTypeMap {
		if strings.HasPrefix(desc.telnet.termType, n) {
			desc.telnet.options = item
			if item.CharMap != nil {
				desc.telnet.charMap = item.CharMap
			}
		} else if strings.HasSuffix(desc.telnet.termType, n) {
			desc.telnet.options = item
			if item.CharMap != nil {
				desc.telnet.charMap = item.CharMap
			}
		}
	}
}

func (desc *descData) getCharset() {
	desc.telnet.charset = telSnFilter(string(desc.telnet.subData))
	desc.setCharset()
	errLog("#%v: GOT %v: %v", desc.id, TermOpt2TXT[int(desc.telnet.subType)], desc.telnet.charset)

	desc.sendSub(desc.telnet.charset, TermOpt_CHARSET, SB_ACCEPTED)
}

func (desc *descData) captureSubSeqData(data byte) {

	if desc.telnet.subLength > maxSubLen {
		desc.telnet.subMode = false
		desc.telnet.subData = []byte{}
		desc.telnet.subLength = 0
		errLog("#%v subsequece size went over %v, abort", desc.id, maxSubLen)
		return
	}

	//7-bit ascii only
	if data >= ' ' && data <= '~' {
		desc.telnet.subData = append(desc.telnet.subData, data)
	}
}

func (desc *descData) ingestLine() {
	desc.inputLock.Lock()

	//Too many lines
	if desc.numLines > maxLines {
		desc.inputLock.Unlock()
		desc.inputFull()
		return
	}
	//Append line to buffer
	var buf string
	if !desc.telnet.options.UTF {
		buf = encodeToUTF(desc.telnet.charMap, desc.inputBuffer)
	} else {
		buf = string(desc.inputBuffer)
	}
	desc.lineBuffer = append(desc.lineBuffer, buf)
	desc.numLines++

	if desc.inputBufferLen != 0 {
		mudLog("#%v: %v: %v", desc.id, desc.cAddr, buf)
	}

	//Reset input buffer
	desc.inputBuffer = []byte{}
	desc.inputBufferLen = 0

	desc.inputLock.Unlock()
}

func (desc *descData) handleTelCmd(command, option byte) {
	//Sub-negotation sequence START
	if command == TermCmd_SB {

		desc.telnet.subData = []byte{}
		desc.telnet.subMode = true
		desc.telnet.subType = option

	} else if command == TermCmd_WILL {

		if option == TermOpt_TERMINAL_TYPE {
			desc.sendSub("", TermOpt_TERMINAL_TYPE, SB_SEND)
		} else if option == TermOpt_SUP_GOAHEAD {
			desc.telnet.options.SUPGA = true
		}

	} else if command == TermCmd_DO {

		if option == TermOpt_CHARSET {
			desc.sendSub(charsetSend, TermOpt_CHARSET, SB_REQ)
		} else if option == TermOpt_SUP_GOAHEAD {
			desc.telnet.options.SUPGA = true
		}

	} else if command == TermCmd_DONT {

		if option == TermOpt_SUP_GOAHEAD {
			desc.telnet.options.SUPGA = false
		}

	} else if command == TermCmd_WONT {

		if option == TermCmd_GOAHEAD {
			desc.telnet.options.SUPGA = false
		}
	}
}

func (desc *descData) send(format string, args ...any) error {
	var outBytes []byte

	//Format string if args supplied
	var data string
	if args != nil {
		data = fmt.Sprintf(format, args...)
	} else {
		data = format
	}

	//Send telnet go-ahead
	if !desc.telnet.options.SUPGA {
		data = data + string([]byte{TermCmd_IAC, TermCmd_GOAHEAD})
	}

	//Character map translation
	if !desc.telnet.options.UTF {
		outBytes = encodeFromUTF(desc.telnet.charMap, data)
	} else {
		outBytes = []byte(data)
	}

	//Write, check for err or invalid len
	dlen := len(outBytes)
	l, err := desc.conn.Write(outBytes)
	if err != nil || dlen != l {
		mudLog("#%v: %v: write failed (connection lost)", desc.id, desc.cAddr)
		return err
	}

	return nil
}

func (desc *descData) sendln(format string, args ...any) error {
	return desc.send(format+"\r\n", args...)
}
