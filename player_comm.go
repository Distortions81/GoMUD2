package main

import (
	"strings"
	"time"
)

const MAX_TELLS = 50
const MAX_TELL_LENGTH = 250
const MAX_TELLS_PER_SENDER = 5

type chanData struct {
	flag   Bitmask
	name   string
	cmd    string
	desc   string
	format string
	level  int

	listeners []*characterData
}

// Do not change channel IDs, max 63
var channels []*chanData = []*chanData{
	0: {name: "Implementor", cmd: "imp", desc: "Implementor chat", format: "[IMP] %v: %v", level: LEVEL_IMPLEMENTOR},
	1: {name: "Administrator", cmd: "admin", desc: "Administrator chat", format: "[ADMIN] %v: %v", level: LEVEL_ADMIN},
	2: {name: "Builder", cmd: "build", desc: "Builder chat", format: "[BUILDER] %v: %v", level: LEVEL_BUILDER},
	3: {name: "Staff", cmd: "staff", desc: "Chat for all staff", format: "[STAFF] %v: %v", level: LEVEL_MODERATOR},
	4: {name: "Moderation", cmd: "mod", desc: "Moderatorion Request", format: "[MOD] %v: %v", level: LEVEL_ANY},
	5: {name: "Announce", cmd: "announce", desc: "Official Announcements", format: "[Announcement] %v: %v", level: LEVEL_ADMIN},
	6: {name: "Congrats", cmd: "grats", desc: "Congratulate someone!", format: "[Grats] %v: %v", level: LEVEL_PLAYER},
	7: {name: "Newbie", cmd: "newb", desc: "A place for newbies to chat or ask for help", format: "[Newbie] %v: %v", level: LEVEL_NEWBIE},
	8: {name: "OOC", cmd: "ooc", desc: "out-of-character chat", format: "[OOC] %v: %v", level: LEVEL_NEWBIE},
}

func sendToChannel(player *characterData, input string, channel int) {
	chd := channels[channel]
	if chd == nil {
		critLog("sendToChannel: Player %v tried to use an invalid chat channel: %v", player.Name, channel)
		return
	}
	if player.Channels.HasFlag(1 << channel) {
		player.send("You currently have the %v channel turned off.", chd.name)
		return
	}
	for _, target := range charList {
		if !target.Channels.HasFlag(1 << channel) {
			if target == player {
				target.send(chd.format, "You", input)
			} else {
				target.send(chd.format, player.Name, input)
			}
		}
	}
}

func cmdChat(player *characterData, input string) {
	cmd := strings.SplitN(input, " ", 2)
	numCmd := len(cmd)

	if numCmd == 0 || input == "" {
		player.send("Syntax: <channel> <message>\r\n\r\nChannel name: command")
		for _, ch := range channels {
			if ch.level > player.Level {
				continue
			}
			player.send("%v: %v", ch.name, ch.cmd)
		}
		return
	}

	if numCmd < 2 {
		player.send("But what do you want to say?")
		return
	}

	for c, ch := range channels {
		if strings.EqualFold(ch.cmd, cmd[0]) {
			if ch.level > player.Level {
				player.send("You aren't high enough level to use that chat channel.")
				return
			}
			sendToChannel(player, cmd[1], c)
			return
		}
	}
	for c, ch := range channels {
		if strings.HasPrefix(ch.cmd, cmd[0]) {
			if ch.level > player.Level {
				player.send("You aren't high enough level to use that chat channel.")
				return
			}
			sendToChannel(player, cmd[1], c)
			return
		}
	}
	player.send("That doesn't seem to be a valid channel.")
}

func cmdChannels(player *characterData, input string) {
	if input == "" {
		player.send("channel command: (on/off) channel name")
		for c, ch := range channels {
			var status string
			if player.Channels.HasFlag(1 << c) {
				status = "OFF"
			} else {
				status = "ON"
			}
			player.send("%10v: (%3v) %v", ch.cmd, status, ch.name)
		}
		player.send("\r\n<channel command> (toggles on/off)")
		return
	}

	for c, ch := range channels {
		if ch.cmd == strings.ToLower(input) {
			if player.Channels.HasFlag(1 << c) {
				player.Channels.ClearFlag(1 << c)
				player.send("%v channel is now on.", ch.name)
			} else {
				player.Channels.AddFlag(1 << c)
				player.send("%v channel is now off.", ch.name)
			}
			player.dirty = true
		}
	}
	player.send("What channel did you want to toggle?")

}

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
		target.send("%v tells you: %v", player.Name, parts[1])
		player.send("You tell %v: %v", target.Name, parts[1])
		return
	} else if target := checkPlayingMatch(parts[0]); target != nil && len(parts[0]) > 2 {
		if target == player {
			target.send("Talking to yourself?")
			return
		}
		target.send("%v tells you: %v", player.Name, parts[1])
		player.send("You tell %v: %v", target.Name, parts[1])
		return
	}

	if len(parts[1]) > MAX_TELL_LENGTH {
		player.send("That is too long of a message for a tell. Maybe send them some mail.")
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
				player.send("I think that is enough tells for now.")
				return
			}
			target.Tells = append(target.Tells, tellData{SenderName: player.Name, SenderUUID: player.UUID, Message: parts[1], Sent: time.Now().UTC()})
			target.saveCharacter()
			player.send("They are offline at the moment, but you tell was saved.")
			return
		}
	}
	player.send("They don't appear to be online.")
}
