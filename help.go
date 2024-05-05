package main

import (
	"strings"
	"time"
)

var helpPages []helpData

type helpData struct {
	created  time.Time
	modified time.Time

	topic    string
	keywords []string
	authors  []string

	text string
}

func cmdHelp(player *characterData, input string) {

	for _, item := range cmdListStr {
		if strings.EqualFold(strings.TrimSpace(input), item.name) {
			player.send(item.help)
			return
		}
	}

	player.send("Sorry, I didn't find a help page for that.")
}

func loadHelp() {

}

func saveHelp() {

}
