package main

import "strings"

func cmdCinfo(player *characterData, input string) {
	player.send("Characters:")
	for _, item := range charList {
		if item.desc != nil {
			player.send("valid: %v: name: %v id: %v", item.valid, item.Name, item.desc.id)
		} else {
			player.send("valid: %v: name: %v (no link)", item.valid, item.Name)
		}
	}
	player.send("\r\nDescriptors:")
	for _, item := range descList {
		player.send("id: %v, addr: %v, state: %v", item.id, item.cAddr, item.state)
	}
}

func cmdPset(player *characterData, input string) {
	var target *characterData

	args := strings.Split(input, " ")
	numArgs := len(args)
	if numArgs > 0 {
		name := args[0]
		if target = checkPlaying(name); target == nil {
			target.send("They aren't online at the moment.")
		}
		if numArgs > 1 {
			command := strings.ToLower(args[1])
			if command == "level" {

			}
		}
	} else {
		//show options
	}
}
