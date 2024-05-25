package main

import "strings"

// DO NOT CHANGE ORDER
const (
	CONFIG_HIDDEN = 1 << iota
	CONFIG_NOTELL
	CONFIG_NOCHANNEL
	CONFIG_DEAF
	CONFIG_NOWRAP

	//Keep at end, do not use or delete
	CONFIG_MAX
)

type configString struct {
	Name, Description string
}

var configNames map[int]configString = map[int]configString{
	CONFIG_HIDDEN:    {Name: "Hidden", Description: "Don't show up in who, join or leave."},
	CONFIG_NOTELL:    {Name: "NoTell", Description: "Reject tells"},
	CONFIG_NOCHANNEL: {Name: "NoChannel", Description: "Mute all channels"},
	CONFIG_DEAF:      {Name: "Deaf", Description: "Mute say/emote/yell"},
	CONFIG_NOWRAP:    {Name: "NoWrap", Description: "Do not word-wrap"},
}

func cmdConfig(player *characterData, input string) {
	if input == "" {
		for x := 0; ; x++ {
			item := configNames[1<<x]
			if item.Name == "" {
				return
			}
			status := "OFF"
			if player.Config.HasFlag(1 << x) {
				status = "On "
			}
			player.send("%10v: (%v) %v", item.Name, status, item.Description)
		}
	}

	parts := strings.Split(input, " ")
	numParts := len(parts)
	found := false
	for y := 0; y < numParts; y++ {
		for x := 0; ; x++ {
			item := configNames[1<<x]
			if item.Name == "" {
				break
			}
			if strings.EqualFold(item.Name, parts[y]) {
				found = true
				if player.Config.HasFlag(1 << x) {
					player.send("%v is now OFF.", item.Name)
					player.Config.ClearFlag(1 << x)
				} else {
					player.send("%v is now ON", item.Name)
					player.Config.AddFlag(1 << x)
				}
				player.dirty = true
				break
			}
		}
		if !found {
			player.send("I don't see a config option with that name.")
		}
	}
}
