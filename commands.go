package main

func handleCommands(desc *descData, input string) {
	desc.sendToPlaying("%v: %v", desc.character.name, input)
}
