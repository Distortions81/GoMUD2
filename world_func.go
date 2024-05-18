package main

import "fmt"

func lookRoom(player *characterData) {
	buf := ""
	var exitList string
	var exitCount int
	for _, exit := range player.room.Exits {
		if exitCount != 0 {
			exitList = exitList + "{x, "
		}
		exitList = exitList + dirToTextColor[exit.Direction]
		exitCount++
	}

	playersList := ""
	if len(player.room.players) > 0 {
		playersList = "\r\nWho's here: "
		for i, target := range player.room.players {
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
		"\r\n%v:\r\n%v\r\n%vExits: %v{x",
		player.room.Name, player.room.Description,
		playersList, exitList)

	player.send(buf)
}
