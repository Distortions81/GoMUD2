package main

import (
	"strings"
	"time"
)

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
	//mudLog("Player %v added to area/room %v / %v", player.Name, area.Name, room.Name)
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

// Send to all players
func (player *characterData) sendToPlaying(format string, args ...any) {
	for _, target := range charList {
		if target == player {
			continue
		}
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

// Init player, attach descriptor to character, put in saved room.
const ANNOUNCE_LOGIN_REST = time.Minute * 30

func (desc *descData) enterWorld(player *characterData) {

	if !player.Config.hasFlag(CONFIG_HIDDEN) && time.Since(player.SaveTime) > ANNOUNCE_LOGIN_REST {
		player.sendToPlaying("--> %v has returned. <--", player.Name)
	}

	player.valid = true
	player.dirty = true
	desc.character = player
	desc.character.desc = desc
	desc.character.loginTime = time.Now()
	desc.character.idleTime = time.Now()

	charList = append(charList, player)
	player.goTo(player.Loc)
	if !player.Loc.AreaUUID.hasUUID() || !player.Loc.RoomUUID.hasUUID() {
		if sysAreaUUID.hasUUID() && sysRoomUUID.hasUUID() {
			critLog("Fixed %v was in nil area or room.", player.Name)
			player.goTo(LocData{AreaUUID: sysAreaUUID, RoomUUID: sysRoomUUID})
		}
	}

	desc.state = CON_PLAYING
	cmdLook(desc.character, "")
	desc.character.checkTells()
	if player.Level < LEVEL_PLAYER {
		player.send("To see the command list type: HELP COMMANDS")
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

func (player *characterData) quit(doClose bool) {

	player.desc.sendln(fairwellBuf)
	if player.saveCharacter() {
		player.send("Character saved.")
		critLog("Saved %v", player.Name)
	} else {
		player.send("Saving character failed.")
		player.valid = false
		//critLog("Failed to save %v", player.Name)
		return
	}

	if doClose {
		player.desc.state = CON_DISCONNECTED
		player.valid = false
	} else {
		player.send("\r\nChoose a character to play:")
		player.desc.inputLock.Lock()
		player.desc.inputLines = []string{}
		player.desc.numInputLines = 0
		player.valid = false
		player.desc.inputLock.Unlock()
		player.leaveRoom()

		go func(target *characterData) {
			descLock.Lock()
			target.desc.state = CON_CHAR_LIST
			showStatePrompt(target.desc)
			target.valid = false
			descLock.Unlock()
		}(player)

	}
}
