package main

import "time"

type DIR int

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
	Name        string
	Description string

	CreDate time.Time
	ModDate time.Time

	Rooms []roomData
}

type roomData struct {
	Version     int
	Loc         locData
	Name        string
	Description string

	CreDate time.Time
	ModDate time.Time

	Exits []exitData
}

type exitData struct {
	Direction DIR
	Name      string
	ToLoc     locData

	pLink *roomData
}
