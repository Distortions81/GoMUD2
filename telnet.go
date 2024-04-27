package main

import (
	"bufio"
	"net"
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
	mudLog("%v connected.", conn.RemoteAddr().String())
	defer conn.Close()

	sendCommand(conn, TermCmd_DO, TermOpt_SUP_GOAHEAD)
	sendCommand(conn, TermCmd_DO, TermOpt_TERMINAL_TYPE)
	sendCommand(conn, TermCmd_DO, TermOpt_CHARSET)

	// Create a new buffered reader for reading incoming data.
	reader := bufio.NewReader(conn)

	desc := &descData{conn: conn, born: time.Now()}

	// Read incoming data loop.
	for serverState == SERVER_RUNNING {
		// Read a byte.
		data, err := connReadByte(reader, conn)
		if err != nil {
			return
		}

		// Process received data.
		switch data {
		case TermCmd_IAC:
			command, err := connReadByte(reader, conn)
			if err != nil {
				return
			}
			if command == TermCmd_SE {
				desc.telnet.subMode = false
				errLog("sub data: %v: %v", TermOpt2TXT[int(desc.telnet.subType)], string(desc.telnet.subData))
				desc.telnet.subData = []byte{}
				continue
			}

			option, err := connReadByte(reader, conn)
			if err != nil {
				return
			}

			if command == TermCmd_SB && option == TermOpt_TERMINAL_TYPE {
				desc.telnet.subData = []byte{}
				desc.telnet.subMode = true
			}

			if command == TermCmd_WILL {
				if option == TermOpt_TERMINAL_TYPE {
					sendSub(desc, TermOpt_TERMINAL_TYPE, SB_SEND)
				}
			}

			errLog("client: %v %v", TermCmd2Txt[int(command)], TermOpt2TXT[int(option)])
		default:
			if desc.telnet.subMode {
				if desc.telnet.subLength > maxSubLen {
					return
				}
				desc.telnet.subData = append(desc.telnet.subData, data)
			} else {
				if desc.inputBufferLen > maxInputLineLength {
					//Too long of a line
					inputFull(conn)
					return
				}

				if data == '\n' {
					desc.lineBufferLock.Lock()

					//Too lany lines
					if desc.numLines > maxLines {
						inputFull(conn)
						desc.lineBufferLock.Unlock()
						return
					}
					desc.lineBuffer = append(desc.lineBuffer, string(desc.inputBuffer))
					desc.numLines++
					mudLog("%v: %v", conn.RemoteAddr().String(), string(desc.inputBuffer))
					desc.lineBufferLock.Unlock()

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

func sendCommand(conn net.Conn, command, option byte) {
	conn.Write([]byte{TermCmd_IAC, command, option})
	errLog("sent: %v %v", TermCmd2Txt[int(command)], TermOpt2TXT[int(option)])
}

func sendSub(desc *descData, args ...byte) {

	desc.telnet.subType = args[0]
	buf := []byte{TermCmd_IAC, TermCmd_SB}
	buf = append(buf, args...)
	buf = append(buf, []byte{TermCmd_IAC, TermCmd_SE}...)
	desc.conn.Write(buf)

	if len(args) > 1 {
		errLog("sent sub: %v %d", TermOpt2TXT[int(args[0])], args[1])
	}
}

func inputFull(conn net.Conn) {
	buf := "Input buffer full! Closing connection..."
	conn.Write([]byte("\r\n" + buf + "\r\n"))
	mudLog("ERROR: %v: %v", conn.RemoteAddr().String(), buf)
}

func connReadByte(reader *bufio.Reader, conn net.Conn) (byte, error) {
	data, err := reader.ReadByte()
	if err != nil {
		mudLog("%v: Connection closed by server.", conn.RemoteAddr().String())
		return 0, err
	}
	return data, nil
}
