package main

import (
	"strings"
)

const MAX_TELLS = 50
const MAX_TELL_LENGTH = 250
const MAX_TELLS_PER_SENDER = 5

type chanData struct {
	name     string
	cmd      string
	desc     string
	format   string
	level    int
	disabled bool
	special  bool
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
	CHAT_IMP:   {name: "Implementer", cmd: "imp", desc: "Implementer chat", format: "[IMP] %v: %v", level: LEVEL_IMPLEMENTER},
	CHAT_ADMIN: {name: "Administrator", cmd: "admin", desc: "Administrator chat", format: "[ADMIN] %v: %v", level: LEVEL_ADMIN},
	CHAT_BUILD: {name: "Builder", cmd: "build", desc: "Builder chat", format: "[BUILDER] %v: %v", level: LEVEL_BUILDER},
	CHAT_STAFF: {name: "Staff", cmd: "staff", desc: "Chat for all staff", format: "[STAFF] %v: %v", level: LEVEL_MODERATOR},
	CHAT_MOD:   {name: "Moderation", cmd: "mod", desc: "Moderatorion Request", format: "[MOD] %v: %v", level: LEVEL_ANY},
	CHAT_ANN:   {name: "Announce", cmd: "announce", desc: "Official Announcements", format: "[Announcement] %v: %v", level: LEVEL_ADMIN},
	CHAT_GRAT:  {name: "Congrats", cmd: "grats", desc: "Congratulate someone!", format: "[Grats] %v: %v", level: LEVEL_PLAYER},
	CHAT_NEWB:  {name: "Newbie", cmd: "newb", desc: "A place for newbies to chat or ask for help", format: "[Newbie] %v: %v", level: LEVEL_NEWBIE},
	CHAT_OOC:   {name: "OOC", cmd: "ooc", desc: "out-of-character chat", format: "[OOC] %v: %v", level: LEVEL_NEWBIE},
	CHAT_CRAZY: {name: "CrazyTalk", cmd: "crazytalk", desc: "chat with ascii-art text", format: "[Crazy Talk] %v:\r\n%v", level: LEVEL_PLAYER, special: true}, //Has it's own command.
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
	if chd.level > player.Level {
		player.send("Your level isn't high enough to use that channel.")
		return false
	}
	if player.Channels.hasFlag(1 << channel) {
		player.send("You currently have the '%v' (%v) channel turned off.", chd.cmd, chd.name)
		return false
	}
	if player.Config.hasFlag(CONFIG_NOCHANNEL) {
		player.send("You currently have chat channels disabled.")
		return false
	}
	msg := input
	for _, target := range charList {
		if target.Config.hasFlag(CONFIG_NOCHANNEL) {
			continue
		}
		if !target.Channels.hasFlag(1<<channel) && notIgnored(player, target, false) {
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
		//player.send("You currently have channels disabled.")
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
		if ch.level > player.Level {
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
		if ch.level > player.Level {
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
		for c, ch := range channels {
			if ch.disabled {
				continue
			}
			if ch.level > player.Level {
				continue
			}
			status := boolToText(!player.Channels.hasFlag(1 << c))
			player.send("%10v: (%3v) %v", ch.cmd, status, ch.name)
		}
		player.send("\r\n<channel command> (toggles on/off)")
		return
	}

	for c, ch := range channels {
		if ch.disabled {
			continue
		}
		if ch.level > player.Level {
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
