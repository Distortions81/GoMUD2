package main

import (
	"strings"
	"time"
)

func (player *characterData) checkTells() {

	numTells := len(player.Tells)
	if numTells > 0 {
		player.send("{rYou have {R%v {rtells waiting.", numTells-1)
	}
	if player.Config.hasFlag(CONFIG_NOTELL) {
		player.send("You currently have tells disabled!")
	}
}

func cmdTells(player *characterData, input string) {
	if player.Config.hasFlag(CONFIG_NOTELL) {
		player.send("Notice: You currently have tells disabled!")
	}
	numTells := len(player.Tells)
	if numTells > 0 {
		tell := player.Tells[0]
		player.send("At %v"+NEWLINE+"%v told you: %v", tell.Sent.String(), tell.SenderName, tell.Message)
		if numTells > 1 {
			player.Tells = player.Tells[1:numTells]
		} else {
			player.Tells = []tellData{}
		}
		player.dirty = true
	} else {
		player.send("You don't have any tells waiting.")
		return
	}
	if numTells-1 > 0 {
		player.send("You have %v more tells waiting.", numTells-1)
	}
}

func cmdTell(player *characterData, input string) {
	if player.Config.hasFlag(CONFIG_NOTELL) {
		player.send("You currently have tells disabled.")
		return
	}
	parts := strings.SplitN(input, " ", 2)
	partsLen := len(parts)

	if input == "" {
		player.send("Tell whom what?")
		return
	}
	if partsLen == 1 {
		player.send("Tell them what?")
		return
	}

	if target := checkPlaying(parts[0]); target != nil {
		if target == player {
			player.send("You tell yourself: %v", parts[1])
			return
		}
		if target.Config.hasFlag(CONFIG_NOTELL) {
			player.send("Sorry, they have tells disabled.")
		}
		player.send("You tell %v: %v", target.Name, parts[1])
		if notIgnored(player, target, true) {
			target.send("%v tells you: %v", player.Name, parts[1])
		}
		return
	} else if target := checkPlayingPMatch(parts[0]); target != nil && len(parts[0]) > 2 {
		if target == player {
			player.send("You tell yourself: %v", parts[1])
			return
		}
		player.send("You tell %v: %v", target.Name, parts[1])
		if notIgnored(player, target, true) {
			target.send("%v tells you: %v", player.Name, parts[1])
		}
		return
	}

	if len(parts[1]) > MAX_TELL_LENGTH {
		player.send("That is too long of a message for a tell. Maybe mail them a letter?")
	}
	tDesc := descData{valid: true}
	target := tDesc.pLoad(parts[0])
	if target != nil {
		if target.Config.hasFlag(CONFIG_NOTELL) {
			player.send("Sorry, they have tells disabled.")
			return
		}
		if len(target.Tells) < MAX_TELLS {
			var ours int
			for _, tell := range target.Tells {
				if tell.SenderUUID == player.UUID && tell.SenderName == player.Name {
					ours++
				}
			}
			if ours >= MAX_TELLS_PER_SENDER {
				player.send("You've reached the maximum number of offline tells to one person.")
				return
			}
			player.send("You tell %v: %v", target.Name, parts[1])
			player.send("They aren't available right now, but your message has been saved.")
			if notIgnored(player, target, true) {
				target.Tells = append(target.Tells, tellData{SenderName: player.Name, SenderUUID: player.UUID, Message: parts[1], Sent: time.Now().UTC()})
				target.saveCharacter()
			}
			return
		} else {
			player.send("They aren't available right now and they have reached the maxiumum number of stored tells.")
		}
	}
	player.send("I don't see anyone by that name.")
}

func notIgnored(player, target *characterData, feedback bool) bool {
	//Players can't ignore staff.
	if target.Level >= LEVEL_BUILDER && player.Level < LEVEL_BUILDER {
		return true
	}

	for _, item := range target.Ignores {
		if item.Name == player.Name && item.UUID == player.UUID {
			if !item.Silent && feedback {
				player.send("Sorry, %v is ignoring you.", target.Name)
			}
			return false
		}
	}
	return true
}
