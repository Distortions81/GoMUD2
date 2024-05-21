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
}

func cmdTells(player *characterData, input string) {
	numTells := len(player.Tells)
	if numTells > 0 {
		tell := player.Tells[0]
		player.send("At %v\r\n%v told you: %v", tell.Sent.String(), tell.SenderName, tell.Message)
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
			target.send("Talking to yourself?")
			return
		}

		player.send("You tell %v: %v", target.Name, parts[1])
		if notIgnored(player, target) {
			target.send("%v tells you: %v", player.Name, parts[1])
		}
		return
	} else if target := checkPlayingPMatch(parts[0]); target != nil && len(parts[0]) > 2 {
		if target == player {
			target.send("Talking to yourself?")
			return
		}
		player.send("You tell %v: %v", target.Name, parts[1])
		if notIgnored(player, target) {
			target.send("%v tells you: %v", player.Name, parts[1])
		}
		return
	}

	if len(parts[1]) > MAX_TELL_LENGTH {
		player.send("That is too long of a message for a tell. Maybe mail them a letter?")
	}
	tDesc := descData{}
	target := tDesc.pLoad(parts[0])
	if target != nil {
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
			if notIgnored(player, target) {
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

func notIgnored(player, target *characterData) bool {
	for _, item := range target.Ignores {
		if item.Name == player.Name && item.UUID == player.UUID {
			if !item.Silent {
				player.send("Sorry, they are ignoring you.")
			}
			return false
		}
	}
	return true
}
