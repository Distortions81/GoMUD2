package main

import "time"

type DIR int

var areaList map[int]*areaData = make(map[int]*areaData)

func init() {
	sysRooms := make(map[int]*roomData)
	sysRooms[0] = &roomData{
		Version: 1, ID: 0, Name: "The void", Description: "Nothing here."}
	areaList[0] = &areaData{
		Version: 1, ID: 0, Name: "System Area", Rooms: sysRooms}
}

const (
	DIR_NORTH = iota
	DIR_EAST
	DIR_SOUTH
	DIR_WEST
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

type locData struct {
	Area int
	Room int
}

type areaData struct {
	Version     int
	ID          int
	Name        string
	Description string

	CreDate time.Time
	ModDate time.Time

	Rooms map[int]*roomData
}

type roomData struct {
	Version     int
	ID          int
	Loc         locData
	Name        string
	Description string

	CreDate time.Time
	ModDate time.Time

	Players []*characterData
	Exits   []*exitData

	pArea *areaData
}

type exitData struct {
	ExitType  int
	DoorName  string
	Direction DIR
	DirName   string
	ToLoc     locData

	pRoom *roomData
}

func linkAreaPointers() {
	var linkLoops int

	for _, area := range areaList {
		for _, room := range area.Rooms {
			room.pArea = area

			linkLoops++
			/*
				for _, exit := range room.Exits {
					exit.pToLoc
				}
			*/
		}
	}

	errLog("Linked room and exits: %v", linkLoops)
}

var dirToStr [DIR_MAX]string = [DIR_MAX]string{
	DIR_NORTH:  "North",
	DIR_EAST:   "East",
	DIR_SOUTH:  "South",
	DIR_WEST:   "West",
	DIR_DOWN:   "Down",
	DIR_UP:     "Up",
	DIR_CUSTOM: "Custom",
}

var dirToText [DIR_MAX]string = [DIR_MAX]string{
	DIR_NORTH:  "North",
	DIR_EAST:   "East",
	DIR_SOUTH:  "South",
	DIR_WEST:   "West",
	DIR_DOWN:   "Down",
	DIR_UP:     "Up",
	DIR_CUSTOM: "Custom",
}

var dirToTextColor [DIR_MAX]string = [DIR_MAX]string{
	DIR_NORTH:  "{RN{rorth",
	DIR_EAST:   "{GE{gast",
	DIR_SOUTH:  "{BS{bouth",
	DIR_WEST:   "{MW{mest",
	DIR_DOWN:   "{WD{wown",
	DIR_UP:     "{YU{yp",
	DIR_CUSTOM: "{wC{kustom",
}
