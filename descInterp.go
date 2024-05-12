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
			sendCmd(desc.conn, TermCmd_WILL, TermOpt_ECHO)
		}
	} else {
		if desc.telnet.hideEcho {
			desc.telnet.hideEcho = false
			//errLog("#%v No longer suppressing echo for login/pass", desc.id)
			sendCmd(desc.conn, TermCmd_WONT, TermOpt_ECHO)
		}
	}
}

func (player *characterData) listCommands() {
	buf := "\r\nCommands:\r\n"
	for i, item := range cmdListStr {
		if item.cmd.level > player.Level {
			continue
		}
		if i > 0 {
			buf = buf + "\r\n"
		}
		buf = buf + item.help
	}
	player.send(buf)
}

func (player *characterData) handleCommands(input string) {
	cmdStr, args, _ := strings.Cut(input, " ")

	cmdStr = strings.ToLower(cmdStr)
	command := commandList[cmdStr]

	if command != nil {
		if command.checkCommandLevel(player) {
			if command.forceArg != "" {
				command.goDo(player, command.forceArg)
			} else {
				command.goDo(player, args)
			}
		}
	} else {
		//Find best partial match
		var scores map[*commandData]int = make(map[*commandData]int)
		cmdStrLen := len(cmdStr)
		var highScoreCmd *commandData
		var highScore = 0
		for x := 0; x < 2; x++ {
			for c, cmd := range commandList {
				if x == 0 && cmd.noShort {
					continue
				}
				cLen := len(c) - 1

				if cLen < cmdStrLen {
					continue
				}
				for x := 0; x < cmdStrLen; x++ {

					if c[x] == cmdStr[x] {
						scores[cmd]++
						continue
					} else {
						scores[cmd] = 0
						break
					}

				}
				if scores[cmd] > highScore {
					highScore = scores[cmd]
					highScoreCmd = cmd
				}
			}
		}
		if highScore > 0 && highScoreCmd != nil {
			if highScoreCmd.noShort {
				player.send("Did you mean %v? You must type that command in full.", highScoreCmd.name)
				return
			} else {
				if highScoreCmd.goDo != nil {
					if highScoreCmd.checkCommandLevel(player) {
						if highScoreCmd.forceArg != "" {
							highScoreCmd.goDo(player, highScoreCmd.forceArg)
						} else {
							highScoreCmd.goDo(player, args)
						}
					}
					return
				}
			}
		}
		player.send("That isn't an available command.")
		player.listCommands()
	}
}

// Returns true if allowed
func (command *commandData) checkCommandLevel(player *characterData) bool {
	if command != nil && command.level > player.Level {
		player.send("Sorry, you aren't high enough level to use this command.")
		return false
	}
	return true
}
