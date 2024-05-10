package main

import (
	"strings"
	"time"
)

func (player *characterData) send(format string, args ...any) {
	if player.desc == nil {
		return
	}
	player.desc.sendln(format, args...)
}

func (player *characterData) sendToPlaying(format string, args ...any) {
	for _, target := range charList {
		if target == player {
			continue
		}
		target.send(format, args...)
	}
}

func (player *characterData) sendToRoom(format string, args ...any) {
	for _, target := range charList {
		if target == player {
			continue
		}
		target.send(format, args...)
	}
}

func (desc *descData) enterWorld(player *characterData) {
	player.valid = true
	desc.character = player
	desc.character.desc = desc
	desc.character.loginTime = time.Now()
	desc.character.idleTime = time.Now()
	if player.room == nil {
		player.room = areaList[sAreaUUID].Rooms[sRoomUUID]
		player.room.players = append(player.room.players, player)
	}
	charList = append(charList, player)

	desc.state = CON_NEWS
}

func checkPlayingPrint(name string, uuid string) *characterData {
	for _, item := range charList {
		if !item.valid {
			continue
		}
		if item.Name == name || item.UUID == uuid {
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

func (player *characterData) quit(doClose bool) {

	player.desc.sendln(aurevoirBuf)
	player.saveCharacter()
	player.send("Character saved.")

	if doClose {
		player.desc.state = CON_DISCONNECTED
		player.valid = false
	} else {
		player.send("\r\nChoose a character to play:")
		player.desc.inputLock.Lock()
		player.desc.lineBuffer = []string{}
		player.desc.numLines = 0
		player.valid = false
		player.desc.inputLock.Unlock()

		go func(target *characterData) {
			descLock.Lock()
			target.desc.state = CON_CHAR_LIST
			target.valid = false
			gCharList(target.desc)
			descLock.Unlock()
		}(player)

	}
}
