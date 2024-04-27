package main

import (
	"bufio"
	"net"
	"strings"
	"time"
)

const (
	maxInputLineLength = 1024 * 2
	connDeadline       = time.Second * 15
	maxLines           = 50
	maxSubLen          = 40 //rfc930
)

// Handle incoming connections.
func handleConnection(conn net.Conn) {
	defer conn.Close()

	descLock.Lock()
	topID++
	desc := &descData{conn: conn, born: time.Now(), id: topID}
	descList = append(descList, desc)
	descLock.Unlock()

	mudLog("#%v: %v connected.", desc.id, conn.RemoteAddr().String())

	sendCommand(desc, TermCmd_DO, TermOpt_SUP_GOAHEAD)
	sendCommand(desc, TermCmd_DO, TermOpt_TERMINAL_TYPE)
	sendCommand(desc, TermCmd_WILL, TermOpt_CHARSET)
	sendCommand(desc, TermCmd_WILL, TermOpt_SUP_GOAHEAD)

	// Create a new buffered reader for reading incoming data.
	reader := bufio.NewReader(conn)

	_, err := conn.Write(greetBuf)
	if err != nil {
		return
	}
	errLog("#%v: Sent greeting", desc.id)

	// Read incoming data loop.
	for serverState == SERVER_RUNNING {
		// Read a byte.
		data, err := connReadByte(reader, desc)
		if err != nil {
			return
		}

		// Process received data.
		switch data {
		case TermCmd_IAC:
			command, err := connReadByte(reader, desc)
			if err != nil {
				return
			}
			if command == TermCmd_SE {

				if desc.telnet.subType == TermOpt_TERMINAL_TYPE {
					desc.telnet.termType = string(desc.telnet.subData)

					errLog("#%v: GOT %v: %v", desc.id, TermOpt2TXT[int(desc.telnet.subType)], desc.telnet.termType)
				} else if desc.telnet.subType == TermOpt_CHARSET {
					desc.telnet.charset = string(desc.telnet.subData)
					if strings.EqualFold(desc.telnet.charset, "UTF-8") {
						desc.telnet.utf = true
					}
					errLog("#%v: GOT %v: %v", desc.id, TermOpt2TXT[int(desc.telnet.subType)], desc.telnet.charset)

					sendSub(desc, desc.telnet.charset, TermOpt_CHARSET, SB_ACCEPTED)
				} else {
					errLog("#%v: GOT unknown sub data: %v: %v", desc.id, TermOpt2TXT[int(desc.telnet.subType)], string(desc.telnet.subData))
				}

				desc.telnet.subMode = false
				desc.telnet.subData = []byte{}
				continue
			}

			option, err := connReadByte(reader, desc)
			if err != nil {
				return
			}

			if command == TermCmd_SB {
				desc.telnet.subData = []byte{}
				desc.telnet.subMode = true
				desc.telnet.subType = option
			} else if command == TermCmd_WILL {
				if option == TermOpt_TERMINAL_TYPE {
					sendSub(desc, "", TermOpt_TERMINAL_TYPE, SB_SEND)
				}
			} else if command == TermCmd_DO {
				if option == TermOpt_CHARSET {
					sendSub(desc, ";UTF-8;US-ASCII;ASCII", TermOpt_CHARSET, SB_REQ)
				}
			}

			errLog("#%v: Client: %v %v", desc.id, TermCmd2Txt[int(command)], TermOpt2TXT[int(option)])
		default:
			if desc.telnet.subMode {
				if desc.telnet.subLength > maxSubLen {
					return
				}
				desc.telnet.subData = append(desc.telnet.subData, data)
			} else {
				if desc.inputBufferLen > maxInputLineLength {
					//Too long of a line
					inputFull(desc)
					return
				}

				if data == '\n' {
					desc.inputLock.Lock()

					//Too lany lines
					if desc.numLines > maxLines {
						inputFull(desc)
						desc.inputLock.Unlock()
						return
					}
					desc.lineBuffer = append(desc.lineBuffer, string(desc.inputBuffer))
					desc.numLines++
					mudLog("#%v: %v: %v", desc.id, conn.RemoteAddr().String(), string(desc.inputBuffer))
					desc.inputLock.Unlock()

					desc.inputBuffer = []byte{}
					desc.inputBufferLen = 0
					continue
				}

				/* No control chars, no delete, but allow UTF-8 */
				if data >= 32 && data != 127 {
					desc.inputBufferLen += 1
					desc.inputBuffer = append(desc.inputBuffer, data)
				}
			}
		}
	}
}

func sendCommand(desc *descData, command, option byte) {
	desc.conn.Write([]byte{TermCmd_IAC, command, option})
	errLog("#%v: Sent: %v %v", desc.id, TermCmd2Txt[int(command)], TermOpt2TXT[int(option)])
}

func sendSub(desc *descData, data string, args ...byte) {

	buf := []byte{TermCmd_IAC, TermCmd_SB}
	buf = append(buf, args...)
	if data != "" {
		buf = append(buf, []byte(data)...)
	}
	buf = append(buf, []byte{TermCmd_IAC, TermCmd_SE}...)
	desc.conn.Write(buf)

	if len(args) > 1 {
		errLog("#%v: Sent sub: %v %v %d", desc.id, data, TermOpt2TXT[int(args[0])], args[1])
	}
}

func inputFull(desc *descData) {
	buf := "Input buffer full! Closing connection..."
	desc.conn.Write([]byte("\r\n" + buf + "\r\n"))
	mudLog("#%v: ERROR: %v: %v", desc.id, desc.conn.RemoteAddr().String(), buf)
}

func connReadByte(reader *bufio.Reader, desc *descData) (byte, error) {
	data, err := reader.ReadByte()
	if err != nil {
		mudLog("#%v: %v: Connection closed by server.", desc.id, desc.conn.RemoteAddr().String())
		return 0, err
	}
	return data, nil
}
