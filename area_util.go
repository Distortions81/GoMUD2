package main

import (
	"fmt"
	"strings"
)

func cmdLook(player *characterData, input string) {
	if player.Room == nil {
		player.send("You are floating in the void.")
		return
	}
	room := player.Room
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
			"%v:\r\n%v\r\nExits: %v",
			room.Name, room.Description,
			exitList))

	}
}

func cmdGo(player *characterData, input string) {
	input = strings.ToLower(input)
	if player.Room == nil {
		player.send("There is nowhere to go.")
		return
	}
	for _, exit := range player.Room.Exits {
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
}

func (player *characterData) goExit(exit *exitData) {
	if player.Room != nil && exit != nil && exit.pRoom != nil {
		player.Room = exit.pRoom
	}
}

func (player *characterData) printToRoom(buf string) {
	if player.Room != nil && player.Room.Players != nil {
		for _, target := range player.Room.Players {
			target.send(buf)
		}
	}
}
