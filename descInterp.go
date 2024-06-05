package main

import (
	"fmt"
	"strings"
	"time"
)

const LOGIN_IDLE = time.Second * 30
const MENU_IDLE = time.Minute * 10
const CHARACTER_IDLE = time.Minute * 30
const BUILDER_IDLE = time.Hour

func (desc *descData) interp() bool {
	var input string

	desc.inputLock.Lock()

	if desc.numInputLines == 0 {
		//Return if there are no lines
		desc.inputLock.Unlock()
		return false
	} else {
		//Get oldest line
		input = desc.inputLines[0]

		if desc.numInputLines == 1 {
			//If only one line, reset buffer
			desc.inputLines = []string{}
			desc.numInputLines = 0
		} else {
			//otherwise, delete oldest entry and decrement count
			desc.inputLines = desc.inputLines[1:]
			desc.numInputLines--
		}

		desc.inputLock.Unlock()
	}

	desc.idleTime = time.Now()

	//If playing
	if desc.state == CON_PLAYING {
		if desc.character != nil {
			//Run command
			desc.character.idleTime = time.Now()
			desc.character.handleCommands(input)
			mudLog("%v: %v", desc.character.Name, input)
		}
		return true
	}

	//Block empty lines, unless login state is set otherwise
	if input == "" && !loginStateList[desc.state].anyKey {
		//Ignore blank lines, unless set
		return true
	}

	//Run login state function
	if loginStateList[desc.state].goDo != nil {
		loginStateList[desc.state].goDo(desc, input)
	}

	showStatePrompt(desc)

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
	return true
}

func showStatePrompt(desc *descData) {
	//Show prompt from next state
	if loginStateList[desc.state].goPrompt != nil {
		loginStateList[desc.state].goPrompt(desc)
	} else if loginStateList[desc.state].prompt != "" {
		desc.sendln("\r\n" + loginStateList[desc.state].prompt)
	}
}

func (player *characterData) listCommands(input string) {
	if input == "" {
		player.send("Commands:")
	}

	var lastLevel int
	for _, item := range cmdList {
		if input != "" {
			if !strings.EqualFold(input, item.name) {
				continue
			}
		}
		if item.hide {
			continue
		}
		if item.level > player.Level {
			continue
		}

		if lastLevel != item.level {
			player.send("\r\nLevel: %v", levelName[item.level])
			lastLevel = item.level
		}

		var parts string
		for i, arg := range item.args {
			if i > 0 {
				parts += ", "
			}
			parts += fmt.Sprintf("<%v>", arg)
		}
		if parts != "" {
			parts += " "
		}
		player.sendWW("%10v %v-- %v", cEllip(item.name, 10), parts, item.hint)
	}

	if input == "" {
		return
	}
	buf := "\r\nChannels: "
	count := 0
	for _, ch := range channels {
		if ch.disabled {
			continue
		}
		if ch.level > player.Level {
			continue
		}
		if count > 0 {
			buf = buf + ", "
		}
		buf = buf + ch.cmd
		count++
	}
	player.sendWW(buf)
}

func updateOLCHere(player *characterData) {
	if player.OLCEditor.OLCMode != OLC_NONE &&
		player.Config.hasFlag(CONFIG_OLCHERE) {

		player.OLCEditor.Location = player.Loc

		player.OLCEditor.area = player.room.pArea
		player.OLCEditor.room = player.room
	}
}

func (player *characterData) handleCommands(input string) {
	defer reportPanic(player, "command: %v", input)
	defer updateOLCHere(player)

	if !player.Config.hasFlag(CONFIG_OLC) &&
		player.OLCEditor.OLCMode != OLC_NONE {

		args := strings.SplitN(input, " ", 2)
		if input != "" &&
			strings.EqualFold(args[0], "cmd") {
			if len(args) == 2 {
				parseCommand(player, args[1])
			} else {
				parseCommand(player, "")
			}
			return
		}
		interpOLC(player, input)
		return
	}

	parseCommand(player, input)

}

func parseCommand(player *characterData, input string) {
	cmdStr, args, _ := strings.Cut(input, " ")
	cmdStr = strings.ToLower(cmdStr)
	command := cmdMap[cmdStr]
	if command != nil {
		if command.disabled {
			player.send("That command is disabled.")
			return
		}
		if command.checkCommandLevel(player) {
			if command.forceArg != "" {
				command.goDo(player, command.forceArg)
			} else {
				command.goDo(player, args)
			}
		}
	} else {
		if cmdChat(player, input) {
			if !findCommandMatch(cmdList, player, cmdStr, args) {
				player.listCommands("")
			}
		}
	}
}

func findCommandMatch(list []*commandData, player *characterData, cmdStr string, args string) bool {
	//Find best partial match
	var scores map[*commandData]int = make(map[*commandData]int)
	cmdStrLen := len(cmdStr)
	var highScoreCmd *commandData
	var highScore = 0
	for x := 0; x < 2; x++ {
		for _, cmd := range list {
			if cmd.disabled {
				continue
			}
			//Dont match against specific crititcal commands
			if x == 0 && cmd.noShort {
				continue
			}
			c := cmd.name
			cLen := len(c) - 1
			//If command name is shorter, skip
			if cLen < cmdStrLen {
				continue
			}
			//Check if all characters match
			for x := 0; x < cmdStrLen; x++ {

				if c[x] == cmdStr[x] {
					scores[cmd]++
					continue
				} else {
					scores[cmd] = 0
					break
				}

			}
			//Save highest scores
			if scores[cmd] > highScore {
				highScore = scores[cmd]
				highScoreCmd = cmd
			}
		}
	}
	//If we found a match, process
	if highScore > 0 && highScoreCmd != nil {
		if highScoreCmd.noShort {
			//Let player know this command cannot be a partial match
			player.send("Did you mean %v? You must type that command in full.", highScoreCmd.name)
			return true
		} else {
			//Run the command
			if highScoreCmd.goDo != nil {
				if highScoreCmd.checkCommandLevel(player) {
					if highScoreCmd.forceArg != "" {
						highScoreCmd.goDo(player, highScoreCmd.forceArg)
					} else {
						highScoreCmd.goDo(player, args)
					}
				}
				return true
			}
		}
	}
	player.send("That isn't an available command.")
	return false
}

// Returns true if allowed
func (command *commandData) checkCommandLevel(player *characterData) bool {
	if command != nil && command.level > player.Level {
		player.send("Sorry, you aren't high enough level to use the '%v' command.", command.name)
		return false
	}
	return true
}
