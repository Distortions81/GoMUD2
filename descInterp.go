package main

import (
	"strings"
	"time"
)

const LOGIN_AFK = time.Second * 30
const AFK_DESC = time.Minute * 5
const CHARACTER_IDLE = time.Minute * 15

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
			desc.character.idleTime = time.Now()
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
	} else if loginStateList[desc.state].prompt != "" {
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
	buf := "\r\nCommands:\r\n"
	for _, item := range cmdListStr {
		buf = buf + item.help + "\r\n"
	}
	desc.send(buf)
}

func (player *characterData) handleCommands(input string) {
	cmdStr, args, _ := strings.Cut(input, " ")

	cmdStr = strings.ToLower(cmdStr)
	command := commandList[cmdStr]

	if command != nil {
		command.goDo(player, args)
	} else {
		//Find best partial match
		var score map[*commandData]int = make(map[*commandData]int)
		cmdStrLen := len(cmdStr)
		for c, cmd := range commandList {
			if cmd.noShort {
				continue
			}
			cLen := len(c)
			fail := false

			minLen := min(cLen, cmdStrLen)
			for x := 0; x < minLen && !fail; x++ {
				for y := 0; y < minLen && !fail; y++ {
					if c[x] == cmdStr[y] {
						score[cmd]++
					} else {
						score[cmd] = 0
						fail = true
					}
				}
			}
		}
		var bestMatch *commandData
		var highScore int
		for cmd, score := range score {
			if score > highScore {
				highScore = score
				bestMatch = cmd
			}
		}
		if highScore > 0 && bestMatch != nil {
			bestMatch.goDo(player, args)
			return
		}
		player.send("That isn't an available command.")
		cmdListCmds(player.desc)
	}
}
