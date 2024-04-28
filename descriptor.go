package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

const (
	maxInputLineLength = 1024 * 2
	connDeadline       = time.Second * 15
	maxLines           = 50
	maxSubLen          = 128
)

// Handle incoming connections.
func handleDesc(conn net.Conn, tls bool) {
	var tlsStr string

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
	desc := &descData{
		conn: conn, id: topID, connectTime: time.Now(),
		reader: bufio.NewReader(conn), tls: tls,
		host: hostStr, addr: ipStr, cAddr: cAddr, state: CON_WELCOME}
	descList = append(descList, desc)
	descLock.Unlock()

	//Close desc on return
	defer desc.close()

	if tls {
		tlsStr = " (TLS)"
	}
	mudLog("#%v: %v connected.%v", desc.id, desc.host, tlsStr)

	//Start telnet negotiation
	sendTelnetCmds(desc)

	//Send greeting
	_, err := conn.Write(greetBuf)
	if err != nil {
		return
	}

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
					//Terminal info
					desc.telnet.termType = string(desc.telnet.subData)
					errLog("#%v: GOT %v: %v", desc.id, TermOpt2TXT[int(desc.telnet.subType)], desc.telnet.termType)

					//Charset recieved
				} else if desc.telnet.subType == TermOpt_CHARSET {
					desc.telnet.charset = string(desc.telnet.subData)
					if strings.EqualFold(desc.telnet.charset, "UTF-8") {
						desc.telnet.utf = true
					}
					errLog("#%v: GOT %v: %v", desc.id, TermOpt2TXT[int(desc.telnet.subType)], desc.telnet.charset)

					desc.sendSub(desc.telnet.charset, TermOpt_CHARSET, SB_ACCEPTED)
				} else {
					//Report unexpected reply
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

			//Sub-negotation sequence START
			if command == TermCmd_SB {
				desc.telnet.subData = []byte{}
				desc.telnet.subMode = true
				desc.telnet.subType = option
			} else if command == TermCmd_WILL {
				//Client termType reply
				if option == TermOpt_TERMINAL_TYPE {
					desc.sendSub("", TermOpt_TERMINAL_TYPE, SB_SEND)
				}
			} else if command == TermCmd_DO {
				//Send our charset list
				if option == TermOpt_CHARSET {
					//If we don't get a reply, use this default
					desc.telnet.charset = "ASCII"
					desc.sendSub(";UTF-8;US-ASCII;ASCII", TermOpt_CHARSET, SB_REQ)
				}
			}

			errLog("#%v: Client: %v %v", desc.id, TermCmd2Txt[int(command)], TermOpt2TXT[int(option)])
		default:
			if desc.telnet.subMode {
				//Limit sub-negotation data size, reset
				if desc.telnet.subLength > maxSubLen {
					desc.telnet.subMode = false
					desc.telnet.subData = []byte{}
					desc.telnet.subLength = 0
					return
				}

				desc.telnet.subData = append(desc.telnet.subData, data)
			} else {

				//Limit line length
				if desc.inputBufferLen > maxInputLineLength {
					//Too long of a line
					desc.inputFull()
					return
				}

				//Detect line end
				if data == '\n' || data == '\r' {
					desc.inputLock.Lock()

					//Reject empty line
					if desc.inputBufferLen == 0 {
						desc.inputLock.Unlock()
						continue
					}

					//Too many lines
					if desc.numLines > maxLines {
						desc.inputLock.Unlock()
						desc.inputFull()
						return
					}

					//Append line to buffer
					desc.lineBuffer = append(desc.lineBuffer, string(desc.inputBuffer))
					desc.numLines++
					mudLog("#%v: %v: %v", desc.id, desc.cAddr, string(desc.inputBuffer))

					//Reset input buffer
					desc.inputBuffer = []byte{}
					desc.inputBufferLen = 0

					desc.inputLock.Unlock()
					continue
				}

				//No control chars, no delete, but allow valid UTF-8
				if data >= 32 && data != 127 {
					desc.inputBufferLen += 1
					desc.inputBuffer = append(desc.inputBuffer, data)
				}
			}
		}
	}
}

func (desc *descData) send(format string, args ...any) error {

	//Format string if args supplied
	var data []byte
	if args != nil {
		data = []byte(fmt.Sprintf(format, args...))
	} else {
		data = []byte(format)
	}

	if !desc.telnet.utf {
		data = convertText(charmap.ISO8859_1, data)
	}

	//Write, check for err or invalid len
	dlen := len(data)
	l, err := desc.conn.Write([]byte(data))

	if err != nil || dlen != l {
		mudLog("#%v: %v: write failed (connection lost)", desc.id, desc.cAddr)
		return err
	}

	return nil
}

func convertText(charmap *charmap.Charmap, data []byte) []byte {
	var tmp []byte
	for _, myRune := range data {

		enc := charmap.NewEncoder()
		win, err := enc.String(string(myRune))
		if err != nil {
			tmp = append(tmp, []byte("?")...)
			continue
		}
		tmp = append(tmp, []byte(win)...)
	}
	return tmp
}
