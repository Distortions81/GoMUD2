package main

type FLAG uint64

const (
	CHAN_IMP = 1 << iota
	CHAN_BUILDER
	CHAN_NEWBIE
	CHAN_ANNOUNCE
	CHAN_GRATS
	CHAN_MODERATION

	//Keep at end, do not use
	CHAN_MAX
)

type chanData struct {
	Flag   int
	Name   string
	Short  string
	Desc   string
	Format string
	Level  int

	listeners []*characterData
}

var channels []chanData = []chanData{
	1: {Name: "Implementor", Short: "Imp", Desc: "Implementor chat", Format: "[IMP] %v says: %v", Level: LEVEL_IMPLEMENTOR},
	2: {Name: "Administrator", Short: "Admin", Desc: "Administrator chat", Format: "[ADMIN] %v says: %v", Level: LEVEL_ADMIN},
	3: {Name: "Builder", Short: "Bld", Desc: "Builder chat", Format: "[BLD] %v says: %v", Level: LEVEL_BUILDER},
	4: {Name: "Moderation", Short: "Mod", Desc: "Moderatorion Request", Format: "[MOD] %v says: %v", Level: LEVEL_ANY},
	5: {Name: "Announce", Short: "Announce", Desc: "Announcements", Format: "[Announcements] %v says: %v", Level: LEVEL_ADMIN},
	6: {Name: "Congrats", Short: "Grats", Desc: "Congratulations", Format: "[Grats] %v says: %v", Level: LEVEL_PLAYER},
	7: {Name: "Newbie", Short: "Newb", Desc: "Newbie chat", Format: "[Newbie] %v says: %v", Level: LEVEL_NEWBIE},
}
