package main

import "strings"

func handleCommands(desc *descData, input string) {
	trimInput := strings.TrimSpace(input)

	if strings.EqualFold(trimInput, desc.lastChat) {
		desc.send("Stop repeating yourself please.")
		return
	}
	desc.lastChat = trimInput
	desc.sendToPlaying("%v: %v", desc.character.name, input)
}
