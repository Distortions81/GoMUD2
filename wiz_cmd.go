package main

import (
	"strconv"
	"strings"
)

func cmdDisable(player *characterData, input string) {
	parts := strings.SplitAfterN(input, " ", 2)
	numParts := len(parts)

	if input == "" {
		player.send("Disable a command or a channel?")
		return
	}
	if numParts == 1 {
		player.send("disable which item?")
		return
	}
	if numParts == 2 {
		if strings.EqualFold(parts[0], "command") {
			for _, cmd := range commandList {
				if strings.EqualFold(cmd.name, parts[1]) {
					cmd.disable = true
					//write disabled status here
				}
			}
		} else if strings.EqualFold(parts[0], "channel") {
			for _, ch := range channels {
				if strings.EqualFold(ch.cmd, parts[1]) {
					ch.disabled = true
					//write disabled status here
				}
			}
		}
	}
}

func cmdConInfo(player *characterData, input string) {
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

	if input == "" {
		cmdHelp(player, "pset")
	}

	args := strings.Split(input, " ")
	numArgs := len(args)

	name := args[0]
	if target = checkPlaying(name); target == nil {
		player.send("They aren't online at the moment.")
		return
	}

	if numArgs < 2 {
		cmdHelp(player, "pset")
		return
	}

	command := strings.ToLower(args[1])
	if command == "level" {
		if numArgs < 3 {
			cmdHelp(player, "pset")
			return
		}
		level, err := strconv.Atoi(args[2])
		if err != nil {
			player.send("That isn't a number.")
			return
		} else {
			if level > player.Level {
				player.send("You can't set a player's level to a level higher than your own.")
				return
			}
			target.Level = level
			player.send("%v's level is now %v.", target.Name, target.Level)
			target.dirty = true
			return
		}
	}

}

func cmdTransport(player *characterData, input string) {
	cmd := strings.SplitN(input, " ", 2)
	cmdLen := len(cmd)

	if input == "" {
		player.send("Send who where?")
		return
	}
	if cmdLen == 1 {
		player.send("Send them where?")
	}
	if cmdLen == 2 {
		if target := checkPlayingPMatch(cmd[0]); target != nil {
			target.leaveRoom()
			target.send("You have been forced to recall.")
			target.goTo(LocData{AreaUUID: sysAreaUUID, RoomUUID: sysRoomUUID})
			player.send("Sent %v to %v.")
			return
		}
		player.send("I don't see anyone by that name.")
	}
}
