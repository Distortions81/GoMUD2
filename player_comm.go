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
var channels []chanData = []chanData{
	0: {Name: "Implementor", Short: "Imp", Desc: "Implementor chat", Format: "[IMP] %v says: %v", Level: LEVEL_IMPLEMENTOR},
	1: {Name: "Administrator", Short: "Admin", Desc: "Administrator chat", Format: "[ADMIN] %v says: %v", Level: LEVEL_ADMIN},
	2: {Name: "Builder", Short: "Build", Desc: "Builder chat", Format: "[BUILDER] %v says: %v", Level: LEVEL_BUILDER},
	3: {Name: "Staff", Short: "Staff", Desc: "Chat for all staff", Format: "[STAFF] %v says: %v", Level: LEVEL_MODERATOR},
	4: {Name: "Moderation", Short: "Mod", Desc: "Moderatorion Request", Format: "[MOD] %v says: %v", Level: LEVEL_ANY},
	5: {Name: "Announce", Short: "Announce", Desc: "Official Announcements", Format: "[Announcement] %v says: %v", Level: LEVEL_ADMIN},
	6: {Name: "Congrats", Short: "Grats", Desc: "Congratulate someone!", Format: "[Grats] %v says: %v", Level: LEVEL_PLAYER},
	7: {Name: "Newbie", Short: "Newb", Desc: "A place for newbies to chat or ask for help", Format: "[Newbie] %v says: %v", Level: LEVEL_NEWBIE},
}