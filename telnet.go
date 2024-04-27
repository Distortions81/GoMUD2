package main

import (
	"bufio"
	"fmt"
	"net"
)

var (
	subType byte
	subMode bool
	subData []byte
)

// Handle incoming connections.
func handleConnection(conn net.Conn) {
	defer conn.Close()

	sendCommand(conn, TermCmd_DO, TermOpt_SUP_GOAHEAD)
	sendCommand(conn, TermCmd_DO, TermOpt_TERMINAL_TYPE)
	sendCommand(conn, TermCmd_DO, TermOpt_CHARSET)

	// Create a new buffered reader for reading incoming data.
	reader := bufio.NewReader(conn)

	// Read incoming data loop.
	for {
		// Read a byte.
		data, err := reader.ReadByte()
		if err != nil {
			fmt.Println("Connection closed by server.")
			return
		}

		// Process received data.
		switch data {
		case TermCmd_IAC:
			command, err := reader.ReadByte()
			if err != nil {
				fmt.Println("Connection closed by server.")
				return
			}
			if command == TermCmd_SE {
				subMode = false
				fmt.Printf("sub data: %v: %v\n", TermOpt2TXT[int(subType)], string(subData))
				subData = []byte{}
				continue
			}

			option, err := reader.ReadByte()
			if err != nil {
				fmt.Println("Connection closed by server.")
				return
			}

			if command == TermCmd_SB && option == TermOpt_TERMINAL_TYPE {
				subData = []byte{}
				subMode = true
			}

			if command == TermCmd_WILL {
				if option == TermOpt_TERMINAL_TYPE {
					sendSub(conn, TermOpt_TERMINAL_TYPE, SB_SEND)
				}
			}

			fmt.Printf("client: %v %v\n", TermCmd2Txt[int(command)], TermOpt2TXT[int(option)])
		default:
			if !subMode {
				fmt.Print(string(data))
			} else {
				subData = append(subData, data)
			}
		}
	}
}

func sendCommand(conn net.Conn, command, option byte) {
	conn.Write([]byte{TermCmd_IAC, command, option})
	fmt.Printf("sent: %v %v\n", TermCmd2Txt[int(command)], TermOpt2TXT[int(option)])
}

func sendSub(conn net.Conn, args ...byte) {

	subType = args[0]
	buf := []byte{TermCmd_IAC, TermCmd_SB}
	buf = append(buf, args...)
	buf = append(buf, []byte{TermCmd_IAC, TermCmd_SE}...)
	conn.Write(buf)

	if len(args) > 1 {
		fmt.Print("sent sub: ")
		fmt.Printf("%v %d", TermOpt2TXT[int(args[0])], args[1])
		fmt.Println()
	}
}
