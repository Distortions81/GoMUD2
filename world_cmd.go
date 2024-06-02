package main

import (
	"strings"
)

func cmdLook(player *characterData, input string) {
	if player.room == nil {
		player.send("You are floating in the nil.")
		return
	}

	if input == "" {
		lookRoom(player)
	} else {
		player.send("Who? What? Huh?")
	}

}

func cmdGo(player *characterData, input string) {
	input = strings.ToLower(input)
	if player.room == nil {
		player.send("There is nowhere to go.")
		return
	}
	for _, exit := range player.room.Exits {
		if exit.Direction == DIR_CUSTOM {
			if strings.HasPrefix(strings.ToLower(exit.DirName), input) {
				player.send("You go %v", exit.DirName)
				player.goExit(exit)
				cmdLook(player, "")
				return
			}
		} else {
			dirStr := dirToText[exit.Direction]
			dirName := strings.ToLower(dirStr)
			if strings.HasPrefix(dirName, input) {
				player.send("You go %v{x", dirToTextColor[exit.Direction])
				player.goExit(exit)
				cmdLook(player, "")
				return
			}
		}
	}
	player.send("Go where?")
}

func (player *characterData) goExit(exit *exitData) {
	if player.room != nil && exit != nil && exit.pRoom != nil {
		var dirStr string
		if exit.Direction == DIR_CUSTOM {
			dirStr = exit.DirName
		} else {
			dirStr = dirToTextColor[exit.Direction]
		}
		player.sendToRoom("%v leaves %v{x", player.Name, dirStr)
		player.leaveRoom()
		player.room = exit.pRoom
		player.room.players = append(player.room.players, player)

		player.Loc = exit.ToRoom
		if exit.Direction == DIR_CUSTOM {
			dirStr = exit.DirName
		} else {
			newDir := exit.Direction.revDir()
			dirStr = dirToTextColor[newDir]
		}
		player.sendToRoom("%v arrives from the %v{x", player.Name, dirStr)
		player.dirty = true
	}
}
