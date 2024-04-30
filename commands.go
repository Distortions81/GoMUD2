package main

import "strings"

const MAX_CHAT_REPEAT = 5

func handleCommands(desc *descData, input string) {
	trimInput := strings.TrimSpace(input)

	if strings.EqualFold(trimInput, desc.lastChat) {
		desc.chatRepeatCount++
		if desc.chatRepeatCount >= MAX_CHAT_REPEAT {
			desc.send("Stop repeating yourself please.")
			return
		}
	}
	desc.lastChat = trimInput

	desc.sendToPlaying("%v: %v", desc.character.name, input)
}
