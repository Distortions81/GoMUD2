package main

import "time"

type DIR int

var areaList map[int]*areaData = make(map[int]*areaData)

func init() {
	sysRooms := make(map[int]*roomData)
	sysRooms[1] = &roomData{
		Version: 1, ID: 1, Loc: locData{Area: 1, Room: 1}, Name: "The void", Description: "Nothing here."}
	areaList[1] = &areaData{
		Version: 1, ID: 1, Name: "System Area", Rooms: sysRooms}
}

const (
	DIR_NORTH = iota
	DIR_EAST
	DIR_SOUTH
	DIR_WEST
	DIR_DOWN
	DIR_UP

	//Keep at end
	DIR_MAX
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

	Exits []exitData

	pArea *areaData
}

type exitData struct {
	Direction DIR
	Name      string
	ToLoc     locData

	pToLoc *roomData
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
