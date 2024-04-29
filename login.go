package main

const (
	PASSWORD_HASH_COST     = 10
	MAX_PLAYER_NAME_LENGTH = 32
	MIN_PLAYER_NAME_LENGTH = 2
)

func handleCommands(desc *descData, input string) {
	desc.sendln("Echo: %v", input)
}
