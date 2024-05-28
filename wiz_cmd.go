package main

import (
	"fmt"
	"goMUD2/figletlib"
	"strconv"
	"strings"
)

const disCol = 4

func cmdBoom(player *characterData, input string) {
	buf := fmt.Sprintf("%v booms: %v", player.Name, input)
	boom := figletlib.TXTToAscii(buf)
	for _, target := range charList {
		target.send(boom)
	}
}

func cmdConInfo(player *characterData, input string) {
	player.send("Descriptors:")
	for _, item := range descList {
		player.send("\r\nID: %-32v IP: %v", item.id, item.ip)
		player.send("State: %-29v DNS: %v", stateName[item.state], item.dns)
		if item.character != nil {
			player.send("Name: %-30v Account: %v", item.character.Name, item.account.Login)
		}
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
