package main

import (
	"time"
)

func (player *characterData) send(format string, args ...any) {
	if player.desc == nil {
		return
	}
	player.desc.sendln(format, args...)
}

func (player *characterData) sendToPlaying(format string, args ...any) {
	for _, target := range descList {
		if !target.valid {
			continue
		}
		if target.state == CON_PLAYING {
			target.sendln(format, args...)
		}
	}
}

func (desc *descData) enterWorld(player *characterData) {
	player.valid = true
	desc.character = player
	desc.character.desc = desc
	desc.character.loginTime = time.Now()
	desc.character.idleTime = time.Now()
	characterList = append(characterList, player)
	desc.state = CON_NEWS
	go func(desc *descData) {
		descLock.Lock()
		defer descLock.Unlock()
		desc.character.sendToPlaying("%v fades into view.", desc.character.Name)
	}(desc)
}

func checkPlaying(name string, fingerprint string) *characterData {
	for _, item := range characterList {
		if item.Name == name || item.Fingerprint == fingerprint {
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
			defer descLock.Unlock()
			gCharList(target.desc)
		}(player)
	}
}
