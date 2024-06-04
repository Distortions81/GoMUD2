package main

import (
	"strings"
	"time"
)

func cmdUnban(player *characterData, input string) {
	doBan(player, input, true, false)
}
func cmdBan(player *characterData, input string) {
	doBan(player, input, false, false)
}

func doBan(player *characterData, input string, unban, account bool) {
	args := strings.SplitN(input, " ", 2)
	argCount := len(args)

	if input == "" {
		player.send("Ban a character: ban <player> <reason>")
		return
	}

	reason := "No reason given."
	if argCount == 2 {
		r := strings.TrimSpace(args[1])
		if len(r) > 1 {
			reason = r
		}
	}

	var target *characterData
	if target = checkPlaying(args[0]); target != nil {
		target.send("You have been banned. Reason: %v", reason)
		target.desc.close()
	}

	if target == nil {
		tDesc := descData{}
		if target = tDesc.pLoad(args[0]); target == nil {
			player.send("Unable to find a player by that name.")
			return
		}
	}

	foundBan := false
	if unban {
		for i, item := range target.Banned {
			if !item.Revoked {
				target.Banned[i].Revoked = true
				foundBan = true
			}
		}
		if foundBan {
			player.send("%v has been unbanned.", target.Name)
			critLog("%v has been unbanned by %v.", target.Name, player.Name)

			target.saveCharacter()
			target.valid = false
		} else {
			player.send("%v wasn't banned.", target.Name)
		}
		return
	}

	target.Banned = append(target.Banned, banData{Reason: reason, Date: time.Now().UTC(), BanBy: player.Name})
	target.saveCharacter()
	player.send("%v has been banned: %v", target.Name, reason)
	critLog("%v has been banned: %v: %v", target.Name, player.Name, reason)
	target.quit(true)
	//player.sendToPlaying(" --> %v has been banned. <--", target.Name)
}
