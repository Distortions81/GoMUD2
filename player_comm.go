package main

type chanData struct {
	Flag   Bitmask
	Name   string
	Short  string
	Desc   string
	Format string
	Level  int

	listeners []*characterData
}

// Do not change channel IDs, max 63
var channels []*chanData = []*chanData{
	0: {Name: "Implementor", Short: "Imp", Desc: "Implementor chat", Format: "[IMP] %v: %v", Level: LEVEL_IMPLEMENTOR},
	1: {Name: "Administrator", Short: "Admin", Desc: "Administrator chat", Format: "[ADMIN] %v: %v", Level: LEVEL_ADMIN},
	2: {Name: "Builder", Short: "Build", Desc: "Builder chat", Format: "[BUILDER] %v: %v", Level: LEVEL_BUILDER},
	3: {Name: "Staff", Short: "Staff", Desc: "Chat for all staff", Format: "[STAFF] %v: %v", Level: LEVEL_MODERATOR},
	4: {Name: "Moderation", Short: "Mod", Desc: "Moderatorion Request", Format: "[MOD] %v: %v", Level: LEVEL_ANY},
	5: {Name: "Announce", Short: "Announce", Desc: "Official Announcements", Format: "[Announcement] %v: %v", Level: LEVEL_ADMIN},
	6: {Name: "Congrats", Short: "Grats", Desc: "Congratulate someone!", Format: "[Grats] %v: %v", Level: LEVEL_PLAYER},
	7: {Name: "Newbie", Short: "Newb", Desc: "A place for newbies to chat or ask for help", Format: "[Newbie] %v: %v", Level: LEVEL_NEWBIE},
	8: {Name: "OOC", Short: "OOC", Desc: "out-of-character chat", Format: "[OOC] %v: %v", Level: LEVEL_NEWBIE},
}

func sendToChannel(player *characterData, input string, channel Bitmask) {

	chd := channels[channel]
	if chd == nil {
		critLog("sendToChannel: Player %v tried to use an invalid chat channel: %v", player.Name, channel)
		return
	}
	if player.Channels.HasFlag(channel) {
		player.send("You currently have the %v channel turned off.", chd.Name)
		return
	}
	for _, target := range charList {
		if !target.Channels.HasFlag(channel) {
			if target == player {
				target.send(chd.Format, "You", input)
			} else {
				target.send(chd.Format, player.Name, input)
			}
		}
	}
}

func cmdOOC(player *characterData, input string) {
	sendToChannel(player, input, 8)
}
