package main

import (
	"strings"
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

	//If playing
	if desc.state == CON_PLAYING {
		if desc.character != nil {
			//Run command
			desc.character.handleCommands(input)
			mudLog("%v: %v", desc.character.Name, input)
		}
		return
	}

	//Block empty lines, unless login state is set otherwise
	if input == "" && !loginStateList[desc.state].anyKey {
		//Ignore blank lines, unless set
		return
	}

	//Run login state function
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

func cmdListCmds(desc *descData) {
	desc.sendln("\r\nCommands:\r\n%v", strings.Join(cmdList, "\r\n"))
}

func (player *characterData) handleCommands(input string) {
	cmd, args, _ := strings.Cut(input, " ")

	cmd = strings.ToLower(cmd)
	command := commandList[cmd]

	if command != nil {
		command.goDo(player, args)
	} else {
		cmdListCmds(player.desc)
	}
}
