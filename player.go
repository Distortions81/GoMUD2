package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

// Returns false on error
func (player *characterData) saveCharacter() bool {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if player == nil {
		critLog("savePlayer: Nil player.")
		return false
	} else if player.Fingerprint == "" {
		critLog("savePlayer: Player '%v' doesn't have a fingerprint.", player.Name)
		return false
	}
	player.Version = CHARACTER_VERSION
	player.SaveTime = time.Now()
	fileName := DATA_DIR + ACCOUNT_DIR + player.desc.account.Fingerprint + "/" + player.Fingerprint + ".json"

	err := enc.Encode(&player)
	if err != nil {
		critLog("savePlayer: enc.Encode: %v", err.Error())
		return false
	}

	err = saveFile(fileName, outbuf.Bytes())
	if err != nil {
		critLog("savePlayer: saveFile failed %v", err.Error())
		return false
	}
	player.dirty = false
	return false
}

func (desc *descData) loadCharacter(plrStr string) *characterData {
	if desc == nil || desc.account == nil {
		return nil
	}

	playFingerprint := ""
	for _, target := range desc.account.Characters {
		if target.Login == plrStr {
			playFingerprint = target.Fingerprint
			break
		}
	}
	if playFingerprint == "" {
		errLog("loadPlayer: Player not found in account.")
		return nil
	}

	player := &characterData{}
	target := checkPlaying(plrStr, playFingerprint)

	if target != nil {
		target.send("You are being kicked, someone else is logging in to this character!")
		target.send(aurevoirBuf)
		target.desc.close()

		player = target
	} else {
		data, err := readFile(DATA_DIR + ACCOUNT_DIR + desc.account.Fingerprint + "/" + playFingerprint + ".json")
		if err != nil {
			return nil
		}

		err = json.Unmarshal(data, player)
		if err != nil {
			errLog("loadPlayer: Unable to unmarshal the data.")
			return nil
		}
	}
	return player
}

func (player *characterData) handleCommands(input string) {
	cmd, args, _ := strings.Cut(input, " ")

	cmd = strings.ToLower(cmd)
	command := commandList[cmd]

	if command != nil {
		command.goDo(player, args)
	} else {
		cmdListCmds(player.desc)
	}
}

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

func cmdListCmds(desc *descData) {
	desc.sendln("\r\nCommands:\r\n%v", strings.Join(cmdList, "\r\n"))
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
		player.desc.state = CON_CHAR_LIST
		gCharList(player.desc)
		player.desc.inputLock.Lock()
		player.desc.lineBuffer = []string{}
		player.desc.numLines = 0
		player.desc.inputLock.Unlock()
		player.valid = false
	}
}

func (desc *descData) enterWorld(player *characterData) {
	player.valid = true
	desc.character = player
	desc.character.desc = desc
	desc.character.loginTime = time.Now()
	characterList = append(characterList, player)
	desc.state = CON_NEWS
}

func checkPlaying(name string, fingerprint string) *characterData {
	for _, item := range characterList {
		if item.Name == name || item.Fingerprint == fingerprint {
			return item
		}
	}
	return nil
}
