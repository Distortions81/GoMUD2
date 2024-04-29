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

	if loginStateList[desc.state] != nil {
		loginStateList[desc.state].goDo(desc, input)

		if loginStateList[desc.state] != nil {
			desc.send(loginStateList[desc.state].prompt)
		}
	}
	if desc.state == CON_DISCONNECTED {
		desc.sendln(textFiles["aurevoir"])
		return
	}
}

type loginStates struct {
	prompt string
	goDo   func(desc *descData, input string)
}

var loginStateList = map[int]*loginStates{

	CON_LOGIN: {
		prompt: "To create a new account type NEW\r\nLogin: ",
		goDo:   gLogin,
	},
	CON_PASS: {
		prompt: "Passphrase: ",
		goDo:   gPass,
	},
}

func gLogin(desc *descData, input string) {
	desc.send("login okay.")
	desc.state = CON_PASS
}

func gPass(desc *descData, input string) {
	desc.send("pass okay.")
	desc.state = CON_DISCONNECTED
}
