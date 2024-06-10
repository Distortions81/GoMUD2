package main

import (
	"strings"
)

const MAX_TELLS = 50
const MAX_TELL_LENGTH = 250
const MAX_TELLS_PER_SENDER = 5

type chanData struct {
	name        string
	cmd         string
	desc        string
	format      string
	talkLevel   int
	listenLevel int
	disabled    bool
	special     bool
}

// Do not change order or delete, only append to the list.
const (
	CHAT_IMP = iota
	CHAT_ADMIN
	CHAT_BUILD
	CHAT_STAFF
	CHAT_MOD
	CHAT_ANN
	CHAT_GRAT
	CHAT_NEWB
	CHAT_OOC
	CHAT_CRAZY

	//Do not move, remove or use.
	CHAT_MAX
)

// Do not change channel IDs (MAX 63)
// use 'disabled: true' to disable vs deleting.
// Otherwise a 'new' channel using the old ID will be 'off' if old channel was off for a player.
var channels []*chanData = []*chanData{
	CHAT_IMP: {name: "Implementer", cmd: "imp", desc: "Implementer chat", format: "[IMP] %v: %v",
		talkLevel: LEVEL_IMPLEMENTER, listenLevel: LEVEL_IMPLEMENTER},
	CHAT_ADMIN: {name: "Administrator", cmd: "admin", desc: "Administrator chat", format: "[ADMIN] %v: %v",
		talkLevel: LEVEL_ADMIN, listenLevel: LEVEL_ADMIN},
	CHAT_BUILD: {name: "Builder", cmd: "build", desc: "Builder chat", format: "[BUILDER] %v: %v",
		talkLevel: LEVEL_BUILDER, listenLevel: LEVEL_BUILDER},
	CHAT_STAFF: {name: "Staff", cmd: "staff", desc: "Chat for all staff", format: "[STAFF] %v: %v",
		talkLevel: LEVEL_BUILDER, listenLevel: LEVEL_BUILDER},
	CHAT_MOD: {name: "Moderation", cmd: "mod", desc: "Moderatorion Request", format: "[MOD] %v: %v",
		talkLevel: LEVEL_ANY, listenLevel: LEVEL_BUILDER},
	CHAT_ANN: {name: "Announce", cmd: "announce", desc: "Official Announcements", format: "[Announcement] %v: %v",
		talkLevel: LEVEL_BUILDER, listenLevel: LEVEL_ANY},
	CHAT_GRAT: {name: "Congrats", cmd: "grats", desc: "Congratulate someone!", format: "[Grats] %v: %v",
		talkLevel: LEVEL_PLAYER, listenLevel: LEVEL_ANY},
	CHAT_NEWB: {name: "Newbie", cmd: "newb", desc: "A place for newbies to chat or ask for help", format: "[Newbie] %v: %v",
		talkLevel: LEVEL_NEWBIE, listenLevel: LEVEL_ANY},
	CHAT_OOC: {name: "OOC", cmd: "ooc", desc: "out-of-character chat", format: "[OOC] %v: %v",
		talkLevel: LEVEL_PLAYER, listenLevel: LEVEL_ANY},
	CHAT_CRAZY: {name: "CrazyTalk", cmd: "crazytalk", desc: "chat with ascii-art text", format: "[Crazy Talk] %v:" + NEWLINE + "%v", talkLevel: LEVEL_PLAYER, listenLevel: LEVEL_ANY, special: true}, //Has it's own command.
}

func sendToChannel(player *characterData, input string, channel int) bool {
	chd := channels[channel]
	if chd == nil {
		critLog("sendToChannel: Player %v tried to use an invalid chat channel: %v", player.Name, channel)
		player.send("Sorry, that isn't a valid chat channel.")
		return false
	}
	if chd.disabled {
		player.send("That channel is disabled.")
		critLog("%v tried to use a disabled comm channel!", player.Name)
		return false
	}
	if chd.talkLevel > player.Level {
		player.send("Your level isn't high enough to speak on that channel.")
		return false
	}
	if chd.listenLevel > player.Level {
		player.send(chd.format, "You", input)
	}
	if player.Channels.hasFlag(1 << channel) {
		player.Channels.addFlag(1 << channel)
		player.send("The %v channel was off, turning it on.", chd.name)
	}
	if player.Config.hasFlag(CONFIG_NOCHANNEL) {
		player.send("NoChannel was enabled, turning it off.")
	}
	msg := input
	for _, target := range charList {
		if chd.listenLevel != LEVEL_ANY && chd.talkLevel < LEVEL_BUILDER &&
			target.Level < chd.listenLevel {
			continue
		}
		if target.Config.hasFlag(CONFIG_NOCHANNEL) {
			continue
		}

		//Bypass channel off for listen-only staff-level channels
		if (chd.talkLevel >= LEVEL_BUILDER && chd.listenLevel == LEVEL_ANY) ||
			//Otherwise, skip if channel is off or player is ignored
			(!target.Channels.hasFlag(1<<channel) && notIgnored(player, target, false)) {
			if channel == CHAT_CRAZY {
				msg = handleCrazy(target, input)
			}
			if target == player {
				target.send(chd.format, "You", msg)
			} else {
				target.send(chd.format, player.Name, msg)
			}
		}
	}
	return true
}

func cmdChat(player *characterData, input string) bool {
	if player.Config.hasFlag(CONFIG_NOCHANNEL) {
		player.send("You currently have channels disabled.")
		return true
	}
	cmd := strings.SplitN(input, " ", 2)
	numCmd := len(cmd)

	if numCmd < 2 {
		//player.send("But what do you want to say?")
		return true
	}
	//Check for full match
	for c, ch := range channels {
		if ch.special {
			continue
		}
		if ch.disabled {
			continue
		}
		if ch.talkLevel > player.Level {
			continue
		}
		if strings.EqualFold(ch.cmd, cmd[0]) {
			if sendToChannel(player, cmd[1], c) {
				return false
			}
		}
	}
	//Otherwise, check for partial match
	for c, ch := range channels {
		if ch.special {
			continue
		}
		if ch.disabled {
			continue
		}
		if ch.talkLevel > player.Level {
			continue
		}
		if strings.HasPrefix(ch.cmd, cmd[0]) {
			sendToChannel(player, cmd[1], c)
			return false
		}
	}
	//player.send("That doesn't seem to be a valid channel.")
	return true
}

func cmdChannels(player *characterData, input string) {
	if player.Config.hasFlag(CONFIG_NOCHANNEL) {
		player.send("You currently have channels disabled.")
		return
	}
	if input == "" {
		player.send("channel command: (on/off) channel name")
		player.send("%10v (%3v)  %v{x", "command:", "on?", "Name")
		for c, ch := range channels {
			status := boolToText(!player.Channels.hasFlag(1 << c))
			cmd := ch.cmd + ":"

			dim := "{W "
			//Disabled
			if ch.disabled {
				status = "{R---{x"
			}
			//No access
			if ch.talkLevel > player.Level &&
				ch.listenLevel > player.Level {
				status = "   "
				cmd = ""
				dim = "{K-"
			}
			//listen only
			if ch.listenLevel <= player.Level &&
				ch.talkLevel > player.Level {
				dim = "{y*{x"
				if ch.talkLevel >= LEVEL_BUILDER {
					status = "{y***{x"
				}
			}
			//Speak only
			if ch.listenLevel > player.Level &&
				ch.talkLevel <= player.Level {
				dim = "{c#{x"
				status = "{c###{x"
			}

			player.send("%10v (%3v) %v%v{x", cmd, status, dim, ch.name)
		}
		player.send(NEWLINE + "<channel command> (toggles on/off)")
		player.send("{y*{x = Listen only, {c#{x = Speak only, {K-{x = No access")
		return
	}

	for c, ch := range channels {
		if ch.listenLevel <= LEVEL_ANY &&
			ch.talkLevel >= LEVEL_BUILDER {
			continue
		}
		if ch.listenLevel > player.Level {
			continue
		}
		if ch.listenLevel > player.Level &&
			ch.talkLevel < player.Level {
			continue
		}
		if strings.EqualFold(ch.cmd, input) {
			if player.Channels.hasFlag(1 << c) {
				player.Channels.clearFlag(1 << c)
				player.send("%v channel is now on.", ch.name)
			} else {
				player.Channels.addFlag(1 << c)
				player.send("%v channel is now off.", ch.name)
			}
			player.dirty = true
			return
		}
	}
	player.send("I can't find a channel by that name.")

}
