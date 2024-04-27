package main

import "fmt"

func (desc *descData) interp() {
	desc.inputLock.Lock()
	defer desc.inputLock.Unlock()

	var input string
	//If there are lines to process
	if desc.numLines > 0 {
		//Get oldest line
		input = desc.lineBuffer[0]

		if desc.numLines == 1 {
			//If only one line, reset buffer
			desc.lineBuffer = []string{}
			desc.numLines = 0
		} else {
			//otherwise, delete oldest entry
			desc.lineBuffer = desc.lineBuffer[1:]
			desc.numLines--
		}
	} else {
		return
	}

	if input == "" {
		return
	}

	fmt.Println(input)

	if desc.state == CON_WELCOME {
		dWelcome(desc)
	}
}

func dWelcome(desc *descData) (fail bool) {
	desc.send("You connected from: %v. Hello: Login: ", desc.conn.RemoteAddr().String())
	return true
}
