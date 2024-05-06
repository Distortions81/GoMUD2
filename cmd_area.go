package main

import "fmt"

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
			exitList = exitList + directionName(exit)
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
