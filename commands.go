package main

func handleCommands(desc *descData, input string) {
	desc.sendln("Echo: %v", input)
}
