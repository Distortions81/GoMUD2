package main

import (
	"sort"
	"strings"
	"time"
)

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
	OLC_NONE:   {name: "NONE"},
	OLC_ROOM:   {name: "ROOM", goDo: olcRoom},
	OLC_AREA:   {name: "AREA", goDo: olcArea},
	OLC_RESET:  {name: "RESET", goDo: olcReset},
	OLC_OBJECT: {name: "OBJECT", goDo: olcObject},
	OLC_MOBILE: {name: "MOBILE", goDo: olcMobile},
}

func cmdOLC(player *characterData, input string) {
	interpOLC(player, input)
}

func olcRoom(player *characterData, input string) {
	if input == "" || strings.EqualFold("help", input) {
		player.send("room edit help goes here")
		return
	} else if strings.EqualFold(input, "desc") {
		//

	} else if strings.EqualFold(input, "list") {
		player.send("Area room list:")
		var roomList []*roomData
		for _, room := range player.room.pArea.Rooms.Data {
			roomList = append(roomList, room)
		}
		sort.Slice(roomList, func(i, j int) bool {
			return roomList[i].VNUM < roomList[j].VNUM || roomList[i].UUID.T < roomList[j].UUID.T || roomList[i].Name < roomList[j].Name
		})
		for _, room := range roomList {
			player.send("VNUM: %-6v Name: %-30v Desc: %-40v", room.VNUM, room.Name, room.Description)
		}
	} else if strings.EqualFold(input, "revnum") {
		var roomList []*roomData
		for _, room := range player.room.pArea.Rooms.Data {
			roomList = append(roomList, room)
		}
		sort.Slice(roomList, func(i, j int) bool {
			return roomList[i].UUID.T < roomList[j].UUID.T || roomList[i].Name < roomList[j].Name
		})
		for r, room := range roomList {
			room.VNUM = r
		}
		player.send("Renumbered %v rooms.", len(roomList))
		player.room.pArea.dirty = true
	}
}

func olcArea(player *characterData, input string) {
	if input == "" || strings.EqualFold("help", input) {
		player.send("area edit help goes here")
		return
	}
}
func olcReset(player *characterData, input string) {
	if input == "" || strings.EqualFold("help", input) {
		player.send("reset edit help goes here")
		return
	}
}
func olcObject(player *characterData, input string) {
	if input == "" || strings.EqualFold("help", input) {
		player.send("object edit help goes here")
		return
	}
}
func olcMobile(player *characterData, input string) {
	if input == "" || strings.EqualFold("help", input) {
		player.send("mobile edit help goes here")
		return
	}
}

func interpOLC(player *characterData, input string) {

	for i, item := range olcModes {
		if strings.EqualFold(item.name, input) {
			player.OLCEditor.OLCMode = i
			player.send("Entering %v edit mode.", item.name)
			item.goDo(player, "")
			return
		}
	}
	if player.OLCEditor.OLCMode != OLC_NONE {
		if strings.EqualFold("exit", input) {
			player.OLCEditor.OLCMode = OLC_NONE
			player.send("Exited OLC editor.")
			return
		}
		for i, item := range olcModes {
			if i == player.OLCEditor.OLCMode {
				item.goDo(player, input)
				return
			}
		}
	}
	player.send("That isn't a valid OLC command.")

	player.send("OLC modes:")
	for _, item := range olcModes {
		if item.goDo == nil {
			continue
		}
		player.send("%-10v", item.name)
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
