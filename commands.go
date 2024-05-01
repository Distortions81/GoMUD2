package main

import (
	"fmt"
	"sort"
	"strings"
)

const WARN_CHAT_REPEAT = 5
const MAX_CHAT_REPEAT = 10

type commandData struct {
	level pLEVEL
	hint  string
	goDo  func(play *playerData, data string)
	args  []string
}

var commandList = map[string]*commandData{
	"say":    {hint: "sends a message", goDo: cmdChat, args: []string{"message"}},
	"quit":   {hint: "quits and disconnects.", goDo: cmdQuit},
	"logout": {hint: "quits back to character selection.", goDo: cmdLogout},
}

var cmdList []string

func init() {
	cmdList = []string{}

	for iName, cmd := range commandList {
		tName := fmt.Sprintf("%15v", iName)
		var buf string
		if cmd.args == nil {
			buf = tName + " -- " + cmd.hint
		} else {
			buf = tName + " -- " + cmd.hint + " : " + iName + " "
			for a, aName := range cmd.args {
				if a > 0 {
					buf = buf + " "
				}
				buf = buf + fmt.Sprintf("<%v>", aName)
			}
		}
		cmdList = append(cmdList, buf)
	}

	sort.Strings(cmdList)
}

func cmdChat(play *playerData, input string) {
	trimInput := strings.TrimSpace(input)
	if strings.EqualFold(trimInput, play.desc.lastChat) {
		play.desc.chatRepeatCount++
		if play.desc.chatRepeatCount >= MAX_CHAT_REPEAT {
			play.desc.close()
			return
		} else if play.desc.chatRepeatCount >= WARN_CHAT_REPEAT {
			play.desc.send("Stop repeating yourself please.")
			return
		}
	}
	play.desc.lastChat = trimInput

	play.sendToPlaying("%v: %v", play.desc.player.name, input)
}

func cmdQuit(play *playerData, input string) {
	play.quit(true)
}

func cmdLogout(play *playerData, input string) {
	play.quit(false)
}

func cmdWho(play *playerData, input string) {
	var buf string
	for _, target := range playList {
		buf = buf + fmt.Sprintf("%30\r\n", target)
	}
	play.send(buf)
}
