package main

import "strings"

func (play *playerData) handleCommands(input string) {
	cmd, args, _ := strings.Cut(input, " ")

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

func (play *playerData) quit(doClose bool) {
	play.desc.send(textFiles["aurevoir"])

	if doClose {
		play.valid = false
		play.desc.close()
	} else {
		play.valid = false
		play.desc.state = CON_WELCOME
	}
}
