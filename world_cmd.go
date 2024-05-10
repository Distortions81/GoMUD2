package main

import (
	"fmt"
	"strings"
)

func cmdLook(player *characterData, input string) {
	if player.room == nil {
		player.send("You are floating in the nil.")
		return
	}
	buf := ""
	room := player.room
	if input == "" {
		var exitList string
		var exitCount int
		for _, exit := range room.Exits {
			if exitCount != 0 {
				exitList = exitList + ", "
			}
			exitList = exitList + dirToTextColor[exit.Direction]
			exitCount++
		}

		playersList := ""
		if len(room.players) > 0 {
			playersList = "\r\nWho's here: "
			for i, target := range room.players {
				if i != 0 {
					playersList = playersList + ", "
				}
				playersList = playersList + target.Name
			}
			if playersList != "" {
				playersList = playersList + "\r\n"
			}
		}

		if exitCount == 0 {
			exitList = "None"
		}
		buf = buf + fmt.Sprintf(
			"\r\n%v:\r\n%v\r\n%vExits: %v",
			room.Name, room.Description,
			playersList, exitList)
	} else {
		buf = "Who? What? Huh?"
	}

	player.send(buf)
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
				player.goExit(exit)
				player.send("You go %v", exit.DirName)
				return
			}
		} else {
			dirStr := dirToStr[exit.Direction]
			if strings.HasPrefix(dirStr, input) {
				player.goExit(exit)
				player.send("You go %v", dirToTextColor[exit.Direction])
				return
			}
		}
	}
	player.send("Go where?")
}

func (player *characterData) goExit(exit *exitData) {
	if player.room != nil && exit != nil && exit.pRoom != nil {
		player.room = exit.pRoom
	}
}

func (player *characterData) printToRoom(buf string) {
	if player.room != nil && player.room.players != nil {
		for _, target := range player.room.players {
			target.send(buf)
		}
	}
}
