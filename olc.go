package main

import (
	"strings"
	"time"
)

func cmdAsave(player *characterData, input string) {
	if player.Level < LEVEL_BUILDER {
		return
	}
	saveAllAreas(false)
	player.send("all areas saved.")
}

func makeRoom(area *areaData) *roomData {
	return &roomData{Version: AREA_VERSION, UUID: makeUUIDString(), Name: "A new room", Description: "Just an empty room", CreDate: time.Now(), ModDate: time.Now(), players: []*characterData{}, Exits: []*exitData{}, pArea: area}
}

// TO DO: currently works from player position, should use different value
// with option of copying player position
func cmdDig(player *characterData, input string) {
	for i, item := range dirToStr {
		if i == DIR_MAX {
			continue
		}
		if strings.EqualFold(item, input) {
			if player.room != nil && player.room.pArea != nil {
				newRoom := makeRoom(player.room.pArea)
				player.room.pArea.Rooms[newRoom.UUID] = newRoom
				player.room.Exits = append(player.room.Exits, &exitData{Direction: DIR(i), pRoom: newRoom})
				newRoom.Exits = append(newRoom.Exits, &exitData{Direction: DIR(i).revDir(), pRoom: player.room})
				player.send("New room created to the: %v", item)
				player.room.pArea.dirty = true
			}
		}
	}
}

func (old DIR) revDir() DIR {
	if old == DIR_NORTH {
		return DIR_SOUTH
	} else if old == DIR_SOUTH {
		return DIR_NORTH
	} else if old == DIR_EAST {
		return DIR_WEST
	} else if old == DIR_WEST {
		return DIR_EAST
	} else if old == DIR_UP {
		return DIR_DOWN
	} else if old == DIR_DOWN {
		return DIR_UP
	} else if old == DIR_NORTH_EAST {
		return DIR_SOUTH_WEST
	} else if old == DIR_SOUTH_WEST {
		return DIR_NORTH_EAST
	} else if old == DIR_NORTH_WEST {
		return DIR_SOUTH_EAST
	} else if old == DIR_SOUTH_EAST {
		return DIR_NORTH_WEST
	} else {
		return old
	}
}
