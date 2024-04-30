package main

func (desc *descData) interp() {
	var input string

	desc.inputLock.Lock()

	if desc.numLines == 0 {
		//Return if there are no lines
		desc.inputLock.Unlock()
		return
	} else {
		//Get oldest line
		input = desc.lineBuffer[0]

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

	//Playing, or disconnected
	if desc.state == CON_PLAYING {
		if input != "" {
			handleCommands(desc, input)
		}
		return
	} else if desc.state == CON_DISCONNECTED {
		desc.sendln(textFiles["aurevoir"])
		return
	}

	//Handle login
	if input == "" && !loginStateList[desc.state].anyKey {
		//Ignore blank lines, unless set
		return
	} else {
		//Otherwise, run the command
		loginStateList[desc.state].goDo(desc, input)
	}

	//Show prompt from next state
	if loginStateList[desc.state].goPrompt != nil {
		loginStateList[desc.state].goPrompt(desc)
	} else {
		desc.send(loginStateList[desc.state].prompt)
	}

}
