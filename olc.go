package main

import (
	"strings"
	"time"
)

func init() {
	for i, item := range olcList {
		item.name = i
	}
}

const (
	OLC_NONE = iota
	OLC_ROOM
	OLC_AREA
	OLC_RESET
	OLC_OBJECT
	OLC_MOBILE

	OLC_MAX
)

type olcModeType struct {
	name string
	goDo func(player *characterData, input string)
}

var olcModes [OLC_MAX]olcModeType = [OLC_MAX]olcModeType{
	OLC_NONE:   {name: "NONE", goDo: olcRoom},
	OLC_ROOM:   {name: "ROOM", goDo: olcRoom},
	OLC_AREA:   {name: "AREA", goDo: olcArea},
	OLC_RESET:  {name: "RESET", goDo: olcReset},
	OLC_OBJECT: {name: "OBJECT", goDo: olcObject},
	OLC_MOBILE: {name: "MOBILE", goDo: olcMobile},
}

var olcList = map[string]*commandData{
	"dig":   {level: LEVEL_BUILDER, hint: "dig out new rooms", goDo: olcDig, args: []string{"direction"}},
	"asave": {level: LEVEL_BUILDER, hint: "force save all areas", goDo: olcAsaveAll},
	"room":  {olcMode: OLC_ROOM, level: LEVEL_BUILDER, hint: "room edit mode", goDo: olcRoom},
}

func cmdOLC(player *characterData, input string) {
	interpOLC(player, input)
}

func olcRoom(player *characterData, input string) {
	if player.OLCEditor.OLCMode != OLC_ROOM {
		player.send("OLC now in room edit mode.")
		player.OLCEditor.OLCMode = OLC_ROOM
	}
}

func olcArea(player *characterData, input string) {
}
func olcReset(player *characterData, input string) {
}
func olcObject(player *characterData, input string) {
}
func olcMobile(player *characterData, input string) {
}

func interpOLC(player *characterData, input string) {
	if player.OLCEditor.OLCMode != OLC_NONE {
		if strings.EqualFold("exit", input) {
			player.OLCEditor.OLCMode = OLC_NONE
			player.send("Exited OLC.")
			return
		}
		for _, item := range olcList {
			if item.olcMode == player.OLCEditor.OLCMode {
				item.goDo(player, input)
				return
			}
		}
	}

	if input == "" && !strings.EqualFold(input, "help") {
		player.send("That doesn't seem to be a OLC command.")
	} else {
		player.send("OLC commands:")
		for _, item := range olcList {
			player.send("%10v -- %v", item.name, item.hint)
		}
	}
}

func olcAsaveAll(player *characterData, input string) {
	if player.Level < LEVEL_BUILDER {
		return
	}
	saveAllAreas(true)
	player.send("all areas saved.")
}

func makeRoom(area *areaData) *roomData {
	return &roomData{Version: ROOM_VERSION, UUID: makeUUID(), Name: "A new room", Description: "Just an empty room", CreDate: time.Now(), ModDate: time.Now(), players: []*characterData{}, Exits: []*exitData{}, pArea: area}
}

// TO DO: currently works from player position, should use different value
// with option of copying player position
func olcDig(player *characterData, input string) {
	for i, item := range dirToText {
		if i == DIR_MAX {
			continue
		}
		if strings.EqualFold(item, input) {
			if player.room != nil && player.room.pArea != nil {
				newRoom := makeRoom(player.room.pArea)
				player.room.pArea.Rooms.Data[newRoom.UUID] = newRoom
				player.room.Exits = append(player.room.Exits,
					&exitData{Direction: DIR(i), pRoom: newRoom,
						ToRoom: LocData{AreaUUID: player.room.pArea.UUID, RoomUUID: newRoom.UUID}})
				newRoom.Exits = append(newRoom.Exits,
					&exitData{Direction: DIR(i).revDir(), pRoom: player.room,
						ToRoom: LocData{AreaUUID: player.room.pArea.UUID, RoomUUID: player.room.UUID}})
				player.send("New room created to the: %v", item)
				player.room.pArea.dirty = true
				return
			}
		}
	}
	player.send("Dig what direction?")
}

func (old DIR) revDir() DIR {

	switch old {
	case DIR_NORTH:
		return DIR_SOUTH
	case DIR_SOUTH:
		return DIR_NORTH
	case DIR_EAST:
		return DIR_WEST
	case DIR_WEST:
		return DIR_EAST
	case DIR_UP:
		return DIR_DOWN
	case DIR_DOWN:
		return DIR_UP
	case DIR_NORTH_EAST:
		return DIR_SOUTH_WEST
	case DIR_SOUTH_WEST:
		return DIR_NORTH_EAST
	case DIR_NORTH_WEST:
		return DIR_SOUTH_EAST
	case DIR_SOUTH_EAST:
		return DIR_NORTH_WEST
	default:
		return old
	}
}
