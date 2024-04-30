package main

import (
	"sort"
	"strings"
)

const WARN_CHAT_REPEAT = 5
const MAX_CHAT_REPEAT = 10

type commandData struct {
	level pLEVEL
	hint  string
	goDo  func(play *playerData, data string)
}

var commandList = map[string]*commandData{
	"say":    {hint: "<message here>", goDo: cmdChat},
	"quit":   {hint: "quits and disconnects.", goDo: cmdQuit},
	"logout": {hint: "quits back to character selection.", goDo: cmdLogout},
}

var cmdList []string

func init() {
	cmdList = []string{}

	for iName, cmd := range commandList {
		cmdList = append(cmdList, iName+" "+cmd.hint)
	}

	sort.Strings(cmdList)
}

func (play *playerData) handleCommands(input string) {
	cmd, args, found := strings.Cut(input, " ")
	if !found {
		play.cmdInvalid()
		return
	}
	cmd = strings.ToLower(cmd)
	command := commandList[cmd]

	if command != nil {
		command.goDo(play, args)
	} else {
		play.cmdInvalid()
	}
}

func (play *playerData) cmdInvalid() {
	play.desc.send("Not a valid command. Commands:\r\n%v", strings.Join(cmdList, "\r\n"))
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

}

func cmdLogout(play *playerData, input string) {

}

func (play *playerData) send(format string, args ...any) {
	if play.desc == nil {
		return
	}
	play.desc.send(format, args...)
}

func (play *playerData) sendToPlaying(format string, args ...any) {
	for _, target := range descList {
		if target.state == CON_PLAYING {
			target.send(format, args...)
		}
	}
}
