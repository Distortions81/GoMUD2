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
		playersList = NEWLINE + "Who's here: "
		for i, target := range player.room.players {

			if i != 0 {
				playersList = playersList + ", "
			}
			playersList = playersList + target.Name
			if target.desc == nil || (target.desc != nil && !target.desc.valid) {
				playersList = playersList + " (no link)"
			}
		}
		if playersList != "" {
			playersList = playersList + NEWLINE
		}
	}

	if exitCount == 0 {
		exitList = "None"
	}
	buf = buf + fmt.Sprintf(
		"%v:"+NEWLINE+"%v"+NEWLINE+"%vExits: %v{x",
		player.room.Name, player.room.Description,
		playersList, exitList)

	player.send(buf)
}
