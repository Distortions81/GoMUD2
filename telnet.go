package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	maxInputLineLength = 1024 * 2
	connDeadline       = time.Second * 15
	maxLines           = 50
	maxSubLen          = 1024
)

// Handle incoming connections.
func handleConnection(conn net.Conn) {
	descLock.Lock()
	topID++
	desc := &descData{conn: conn, id: topID, born: time.Now(), reader: bufio.NewReader(conn)}
	desc.state = CON_WELCOME
	descList = append(descList, desc)
	descLock.Unlock()

	//Close and mark disconnected if we
	defer desc.close()

	mudLog("#%v: %v connected.", desc.id, conn.RemoteAddr().String())

	desc.sendCmd(TermCmd_DO, TermOpt_SUP_GOAHEAD)
	desc.sendCmd(TermCmd_DO, TermOpt_TERMINAL_TYPE)
	desc.sendCmd(TermCmd_WILL, TermOpt_CHARSET)
	desc.sendCmd(TermCmd_WILL, TermOpt_SUP_GOAHEAD)

	_, err := conn.Write(greetBuf)
	if err != nil {
		return
	}

	// Read incoming data loop.
	for serverState == SERVER_RUNNING {
		// Read a byte.
		data, err := desc.readByte()
		if err != nil {
			return
		}

		// Process received data.
		switch data {
		case TermCmd_IAC:
			command, err := desc.readByte()
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

					desc.sendSub(desc.telnet.charset, TermOpt_CHARSET, SB_ACCEPTED)
				} else {
					errLog("#%v: GOT unknown sub data: %v: %v", desc.id, TermOpt2TXT[int(desc.telnet.subType)], string(desc.telnet.subData))
				}

				desc.telnet.subMode = false
				desc.telnet.subData = []byte{}
				continue
			}

			option, err := desc.readByte()
			if err != nil {
				return
			}

			if command == TermCmd_SB {
				desc.telnet.subData = []byte{}
				desc.telnet.subMode = true
				desc.telnet.subType = option
			} else if command == TermCmd_WILL {
				if option == TermOpt_TERMINAL_TYPE {
					desc.sendSub("", TermOpt_TERMINAL_TYPE, SB_SEND)
				}
			} else if command == TermCmd_DO {
				if option == TermOpt_CHARSET {
					desc.sendSub(";UTF-8;US-ASCII;ASCII", TermOpt_CHARSET, SB_REQ)
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
					desc.inputFull()
					return
				}

				if data == '\n' || data == '\r' {
					desc.inputLock.Lock()

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

					desc.lineBuffer = append(desc.lineBuffer, string(desc.inputBuffer))
					desc.numLines++
					mudLog("#%v: %v: %v", desc.id, conn.RemoteAddr().String(), string(desc.inputBuffer))

					desc.inputBuffer = []byte{}
					desc.inputBufferLen = 0

					desc.inputLock.Unlock()
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

func (desc *descData) send(format string, args ...any) error {
	var data []byte
	if args != nil {
		data = []byte(fmt.Sprintf(format, args...))
	} else {
		data = []byte(format)
	}

	dlen := len(data)
	l, err := desc.conn.Write([]byte(data))

	if err != nil || dlen != l {
		mudLog("#%v: %v: write failed (connection lost)", desc.id, desc.conn.RemoteAddr().String())
		return err
	}

	return nil
}

func (desc *descData) sendCmd(command, option byte) error {
	dlen, err := desc.conn.Write([]byte{TermCmd_IAC, command, option})
	if err != nil || dlen != 3 {
		mudLog("#%v: %v: command send failed (connection lost)", desc.id, desc.conn.RemoteAddr().String())
		return err
	}

	errLog("#%v: Sent: %v %v", desc.id, TermCmd2Txt[int(command)], TermOpt2TXT[int(option)])
	return nil
}

func (desc *descData) sendEOR() error {
	dlen, err := desc.conn.Write([]byte{TermOpt_END_OF_RECORD})
	if err != nil || dlen != 1 {
		mudLog("#%v: %v: EOR send failed (connection lost)", desc.id, desc.conn.RemoteAddr().String())
		return err
	}
	errLog("#%v: Sent: %v", desc.id, TermOpt2TXT[int(TermOpt_END_OF_RECORD)])
	return nil
}

func (desc *descData) sendSub(data string, args ...byte) error {
	buf := []byte{TermCmd_IAC, TermCmd_SB}
	buf = append(buf, args...)
	if data != "" {
		buf = append(buf, []byte(data)...)
	}
	buf = append(buf, []byte{TermCmd_IAC, TermCmd_SE}...)
	dlen, err := desc.conn.Write(buf)
	if err != nil || dlen != len(buf) {
		mudLog("#%v: %v: sub send failed (connection lost)", desc.id, desc.conn.RemoteAddr().String())
		return err
	}

	if len(args) > 1 {
		errLog("#%v: Sent sub: %v %v %d", desc.id, data, TermOpt2TXT[int(args[0])], args[1])
	}

	return nil
}

func (desc *descData) inputFull() {
	buf := "Input buffer full! Closing connection..."
	desc.send("\r\n%v\r\n", buf)
	mudLog("#%v: ERROR: %v: %v", desc.id, desc.conn.RemoteAddr().String(), buf)
}

func (desc *descData) readByte() (byte, error) {
	data, err := desc.reader.ReadByte()
	if err != nil {
		mudLog("#%v: %v: Connection closed by server.", desc.id, desc.conn.RemoteAddr().String())
		return 0, err
	}
	return data, nil
}

func (desc *descData) close() {
	if desc == nil {
		return
	}
	desc.state = CON_DISCONNECTED
	if desc.conn != nil {
		desc.conn.Close()
	}
}
