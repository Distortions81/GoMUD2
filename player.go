package main

import (
	"strings"
)

func (target *characterData) quit(disc bool) {
	addRecentLogin(target)

	target.send(fairwellBuf)
	target.sendToRoom("%v slowly fades away.", target.Name)
	target.leaveRoom()
	target.saveCharacter()
	target.valid = false

	if !disc {
		target.send(NEWLINE + "Choose a character to play:")
		target.desc.inputLock.Lock()
		target.desc.inputLines = []string{}
		target.desc.numInputLines = 0
		target.desc.inputLock.Unlock()

		go func(desc *descData) {
			descLock.Lock()
			desc.state = CON_CHAR_LIST
			showStatePrompt(desc)
			descLock.Unlock()
		}(target.desc)
	} else {
		target.desc.kill()
	}
}

// Send player to a room, if they are not in one
func (player *characterData) goTo(loc LocData) {

	area := areaList[loc.AreaUUID]
	if area == nil {
		critLog("Area not found: %v", loc.AreaUUID)
		player.send("That area can't be found: %v", loc.AreaUUID)
		return
	}
	room := area.Rooms.Data[loc.RoomUUID]
	if room == nil {
		critLog("Room not found: %v", loc.RoomUUID)
		player.send("The room %v can't be found in the area: %v", loc.RoomUUID, area.Name)
		return
	}
	player.room = room
	player.Loc = loc

	room.players = append(room.players, player)
	//errLog("Player %v added to area/room %v / %v", player.Name, area.Name, room.Name)
	player.dirty = true

}

// Leave a room, returns false if not found.
func (player *characterData) leaveRoom() bool {
	if player != nil && player.room != nil {
		numPlayers := len(player.room.players)
		if numPlayers == 1 {
			player.room.players = []*characterData{}
			return true
		} else if numPlayers > 1 {
			for c, char := range player.room.players {
				if char.UUID == player.UUID {
					player.room.players = append(player.room.players[:c], player.room.players[c+1:]...)
					return true
				}
			}
		}
		critLog("Failed to remove character %v from room.", player.Name)
		return false
	}
	return false
}

func (player *characterData) send(format string, args ...any) {
	if player.desc == nil {
		return
	}
	player.desc.sendln(format, args...)
}

// Send to all players except self
func (player *characterData) sendToPlaying(format string, args ...any) {
	for _, target := range charList {
		if target == player {
			continue
		}
		target.send(format, args...)
	}
}

// Send to all players
func sendToAll(format string, args ...any) {
	for _, target := range charList {
		target.send(format, args...)
	}
}

// Send message to others in a room
func (player *characterData) sendToRoom(format string, args ...any) {
	if player.room == nil {
		return
	}
	for _, target := range player.room.players {
		if target == player {
			continue
		}
		if target.Config.hasFlag(CONFIG_DEAF) {
			return
		}
		if notIgnored(player, target, false) {
			target.send(format, args...)
		}
	}
}

func checkPlayingUUID(name string, uuid uuidData) *characterData {
	for _, item := range charList {
		if !item.valid {
			continue
		}
		if item.Name == name && item.UUID == uuid {
			return item
		}
	}
	return nil
}

func checkPlaying(name string) *characterData {
	for _, item := range charList {
		if !item.valid {
			continue
		}
		if strings.EqualFold(item.Name, name) {
			return item
		}
	}
	return nil
}

func checkPlayingPMatch(name string) *characterData {
	name = strings.ToLower(name)
	for _, item := range charList {
		if !item.valid {
			continue
		}
		if strings.EqualFold(item.Name, name) {
			return item
		}
	}
	for _, item := range charList {
		if !item.valid {
			continue
		}
		if strings.HasPrefix(strings.ToLower(item.Name), name) {
			return item
		}
	}
	return nil
}
