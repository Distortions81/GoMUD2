package main

import (
	"time"
)

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

	desc.idleTime = time.Now()

	//Playing, or disconnected
	if desc.state == CON_PLAYING {
		if desc.character != nil {
			desc.character.handleCommands(input)
			mudLog("%v: %v", desc.character.Name, input)
		}
		return
	}

	//Handle login
	if input == "" && !loginStateList[desc.state].anyKey {
		//Ignore blank lines, unless set
		return
	}

	if loginStateList[desc.state].goDo != nil {
		loginStateList[desc.state].goDo(desc, input)
	}

	//Show prompt from next state
	if loginStateList[desc.state].goPrompt != nil {
		loginStateList[desc.state].goPrompt(desc)
	} else {
		desc.sendln("\r\n" + loginStateList[desc.state].prompt)
	}

	//Suppress echo for passwords
	if loginStateList[desc.state].hideInfo {
		if !desc.telnet.hideEcho {
			desc.telnet.hideEcho = true
			//errLog("#%v Suppressing echo for login/pass", desc.id)
			desc.sendCmd(TermCmd_WILL, TermOpt_ECHO)
		}
	} else {
		if desc.telnet.hideEcho {
			desc.telnet.hideEcho = false
			//errLog("#%v No longer suppressing echo for login/pass", desc.id)
			desc.sendCmd(TermCmd_WONT, TermOpt_ECHO)
		}
	}
}
