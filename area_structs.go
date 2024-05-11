package main

import "time"

type DIR int

const (
	DIR_NORTH = iota
	DIR_NORTH_EAST
	DIR_EAST
	DIR_SOUTH_EAST
	DIR_SOUTH
	DIR_SOUTH_WEST
	DIR_WEST
	DIR_NORTH_WEST
	DIR_DOWN
	DIR_UP
	DIR_CUSTOM

	//Keep at end
	DIR_MAX
)

const (
	EXIT_NORMAL = 1 << iota
	EXIT_DOOR
	EXIT_HIDDEN
	EXIT_PERSONAL
	EXIT_GUILD
	EXIT_IMMORTAL
	EXIT_KEYED

	//Keep at end
	EXIT_MAX
)

type areaData struct {
	Version     int
	UUID        string
	VNUM        int
	Name        string
	Description string

	CreDate time.Time
	ModDate time.Time

	Rooms map[string]*roomData
	dirty bool
}

type roomData struct {
	Version     int
	UUID        string `json:"-"`
	VNUM        int
	Name        string
	Description string

	CreDate time.Time
	ModDate time.Time

	players []*characterData
	Exits   []*exitData

	pArea *areaData
}

type LocData struct {
	AreaUUID string
	RoomUUID string

	Area, Room int
}

type exitData struct {
	ExitType int
	DoorName string

	Direction DIR
	DirName   string
	ToRoom    LocData

	pRoom *roomData
}

var dirToStr [DIR_MAX]string = [DIR_MAX]string{
	DIR_NORTH:      "North",
	DIR_NORTH_EAST: "Northeast",
	DIR_EAST:       "East",
	DIR_SOUTH_EAST: "Southeast",
	DIR_SOUTH:      "South",
	DIR_SOUTH_WEST: "Southwest",
	DIR_WEST:       "West",
	DIR_NORTH_WEST: "Northwest",
	DIR_DOWN:       "Down",
	DIR_UP:         "Up",

	DIR_CUSTOM: "Custom",
}

var dirToTextColor [DIR_MAX]string = [DIR_MAX]string{
	DIR_NORTH:      "{RN{rorth",
	DIR_NORTH_EAST: "{RN{gE",
	DIR_EAST:       "{GE{gast",
	DIR_SOUTH_EAST: "{BS{GE",
	DIR_SOUTH:      "{BS{bouth",
	DIR_SOUTH_WEST: "{BS{mW",
	DIR_WEST:       "{MW{mest",
	DIR_NORTH_WEST: "{RN{mW",
	DIR_DOWN:       "{CD{cown",
	DIR_UP:         "{YU{yp",

	DIR_CUSTOM: "Custom",
}
