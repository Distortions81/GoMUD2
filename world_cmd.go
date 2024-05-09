package main

import (
	"fmt"
	"strings"
)

func cmdLook(player *characterData, input string) {
	if player.room == nil {
		player.send("You are floating in the void.")
		return
	}
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
		if exitCount == 0 {
			exitList = "None"
		}
		player.send(fmt.Sprintf(
			"\r\n%v:\r\n%v\r\nExits: %v",
			room.Name, room.Description,
			exitList))

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
