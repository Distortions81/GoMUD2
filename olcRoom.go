package main

import (
	"sort"
	"strings"
	"time"
)

var roomCmds []*commandData = []*commandData{
	{name: "dig", goDo: rDig, hint: "Create a new room in <direction>", args: []string{"direction"}},
	{name: "list", goDo: rList, hint: "shows list of rooms in current area"},
	{name: "revnum", goDo: rRevnum, hint: "automatically reassigns new vnums to all room in the area"},
	{name: "description", goDo: rDesc, hint: "Set room description", args: []string{"room description"}},
}

func rDesc(player *characterData, input string) {
}

func rList(player *characterData, input string) {
	player.send("Area room list:")
	var roomList []*roomData
	for _, room := range player.room.pArea.Rooms.Data {
		roomList = append(roomList, room)
	}
	sort.Slice(roomList, func(i, j int) bool {
		return roomList[i].VNUM < roomList[j].VNUM
	})
	for _, room := range roomList {
		player.send("VNUM: %-6v Name: %-30v Desc: %-40v", room.VNUM, room.Name, room.Description)
	}
}
func rRevnum(player *characterData, input string) {
	var roomList []*roomData
	for _, room := range player.room.pArea.Rooms.Data {
		roomList = append(roomList, room)
	}
	//Sort by UUID time
	sort.Slice(roomList, func(i, j int) bool {
		return roomList[i].UUID.T < roomList[j].UUID.T
	})
	for r, room := range roomList {
		room.VNUM = r * vnumSpace
	}
	player.send("Renumbered %v rooms.", len(roomList))
	player.room.pArea.dirty = true
}

func makeRoom(area *areaData) *roomData {
	return &roomData{Version: ROOM_VERSION, UUID: makeUUID(), Name: "A new room", Description: "Just an empty room", CreDate: time.Now(), ModDate: time.Now(), players: []*characterData{}, Exits: []*exitData{}, pArea: area}
}

// TO DO: currently works from player position, should use different value
// with option of copying player position
func rDig(player *characterData, input string) {
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
