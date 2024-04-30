package main

import "strings"

func (play *playerData) handleCommands(input string) {
	cmd, args, found := strings.Cut(input, " ")
	if !found {
		cmdInvalid(play.desc)
		return
	}
	cmd = strings.ToLower(cmd)
	command := commandList[cmd]

	if command != nil {
		command.goDo(play, args)
	} else {
		cmdInvalid(play.desc)
	}
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

func cmdInvalid(desc *descData) {
	desc.send("\r\nCommands:\r\n%v", strings.Join(cmdList, "\r\n"))
}
