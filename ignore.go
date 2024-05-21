package main

import (
	"strings"
	"time"
)

func cmdIgnore(player *characterData, input string) {
	input = strings.ToLower(input)
	parts := strings.SplitN(input, " ", 2)

	if input == "" {
		if len(player.Ignores) == 0 {
			player.send("You aren't currently ignoring anyone. ignore <player name> to add.")
			player.send("For a silent ignore: ignore <player name> silent")
			player.send("Silent ignores don't inform the target that their message was ignored.")
			return
		}
		for _, item := range player.Ignores {
			player.send("%v -- Added: %v", item.Name, item.Added.String())
		}
		return
	}

	arg := ""
	if len(parts) == 2 {
		arg = parts[1]
	}

	if target := checkPlaying(parts[0]); target != nil {
		addIgnore(player, target, arg)
		return
	} else {
		tDesc := descData{}
		if target := tDesc.pLoad(parts[0]); target != nil {
			addIgnore(player, target, arg)
			return
		}
	}

	player.send("Sorry, I don't see a player by that name. Type the full name.")
}

func addIgnore(player, target *characterData, arg string) {
	if target.Level >= LEVEL_MODERATOR {
		player.send("You can't ignore moderators or staff.")
		return
	}
	if target == player {
		player.send("You can't ignore yourself. Maybe consider therapy.")
		return
	}
	doSilent := false
	silentStr := ""
	if strings.EqualFold(arg, "silent") {
		silentStr = " (silent)"
		doSilent = true
	}

	player.dirty = true
	var newIgnores []IgnoreData
	found := false
	for _, item := range player.Ignores {
		if item.Name == target.Name && item.UUID == target.UUID {
			player.send("%v is no longer ignored.", target.Name)
			found = true
			continue
		}
		newIgnores = append(newIgnores, item)
	}
	if found {
		player.Ignores = newIgnores
		return
	}
	player.Ignores = append(player.Ignores, IgnoreData{Name: target.Name, UUID: target.UUID, Silent: doSilent, Added: time.Now().UTC()})

	player.send("%v has been added to your ignore list.%v", target.Name, silentStr)
}
