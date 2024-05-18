package main

import (
	"strings"
	"time"
)

var olcList = map[string]*commandData{
	"dig":   {level: LEVEL_BUILDER, hint: "dig out new rooms", goDo: cmdDig, args: []string{"direction"}},
	"asave": {level: LEVEL_BUILDER, hint: "force save all areas", goDo: cmdAsaveAll},
	"room":  {level: LEVEL_BUILDER, hint: "room edit mode", goDo: cmdRoom},
}

func cmdRoom(player *characterData, input string) {
	player.OLCMode = OLC_ROOM

	if strings.EqualFold(input, "exit") {
		player.OLCMode = OLC_NONE
	}
}

func interpOLC(player *characterData, input string) {
	input = strings.ToLower(input)
	args := strings.SplitN(input, " ", 2)
	numArgs := len(args)
	if numArgs == 0 {
		return
	}
	olcCmd := olcList[args[0]]
	if olcCmd != nil {
		if numArgs > 1 {
			olcCmd.goDo(player, args[1])
		} else {
			olcCmd.goDo(player, "")
		}
		return
	}
	if input != "" && !strings.EqualFold(input, "help") {
		player.send("That doesn't seem to be a OLC command.")
	} else {
		player.send("OLC commands:")
	}
	for i, item := range olcList {
		player.send("%10v -- %v", i, item.hint)
	}
}

func cmdAsaveAll(player *characterData, input string) {
	if player.Level < LEVEL_BUILDER {
		return
	}
	saveAllAreas(true)
	player.send("all areas saved.")
}

func makeRoom(area *areaData) *roomData {
	return &roomData{Version: ROOM_VERSION, UUID: makeUUIDString(), Name: "A new room", Description: "Just an empty room", CreDate: time.Now(), ModDate: time.Now(), players: []*characterData{}, Exits: []*exitData{}, pArea: area}
}

// TO DO: currently works from player position, should use different value
// with option of copying player position
func cmdDig(player *characterData, input string) {
	for i, item := range dirToText {
		if i == DIR_MAX {
			continue
		}
		if strings.EqualFold(item, input) {
			if player.room != nil && player.room.pArea != nil {
				newRoom := makeRoom(player.room.pArea)
				player.room.pArea.Rooms[newRoom.UUID] = newRoom
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
