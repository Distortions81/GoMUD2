package main

func (desc *descData) interp() {
	var input string

	desc.inputLock.Lock()
	if desc.numLines == 0 {
		desc.inputLock.Unlock()
		return
	} else {
		//Get oldest line
		input = desc.lineBuffer[0]
		if input == "" {
			desc.inputLock.Unlock()
			return
		}

		if desc.numLines == 1 {
			//If only one line, reset buffer
			desc.lineBuffer = []string{}
			desc.numLines = 0
		} else {
			//otherwise, delete oldest entry and decrement count
			desc.lineBuffer = desc.lineBuffer[1:]
			desc.numLines--
		}
		desc.inputLock.Unlock()
	}

	if desc.state == CON_WELCOME {
		dWelcome(desc, input)
	}
}

func dWelcome(desc *descData, input string) (fail bool) {
	desc.send("You connected from: %v. Hello: Login: ", desc.conn.RemoteAddr().String())
	return true
}
