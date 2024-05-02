package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

// Returns false on error
func (player *playerData) savePlayer() bool {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if player == nil {
		critLog("Nil player.")
		return false
	} else if player.Fingerprint == "" {
		critLog("Player '%v' doesn't have a fingerprint.", player.Name)
		return false
	}
	player.Version = PLAYER_VERSION
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

func (desc *descData) loadPlayer(plrStr string) bool {
	if desc == nil || desc.account == nil {
		return false
	}

	playFingerprint := ""
	for _, target := range desc.account.Characters {
		if target.Login == plrStr {
			playFingerprint = target.Fingerprint
			break
		}
	}
	if playFingerprint == "" {
		errLog("Player not found in account.")
		return false
	}

	data, err := readFile(DATA_DIR + ACCOUNT_DIR + desc.account.Fingerprint + "/" + playFingerprint + ".json")
	if err != nil {
		return false
	}

	player := &playerData{}
	err = json.Unmarshal(data, player)
	if err != nil {
		errLog("loadPlayer: Unable to unmarshal the data.")
		return false
	}
	player.valid = true

	desc.player = player
	desc.player.desc = desc

	playList = append(playList, player)
	return true
}

func (play *playerData) handleCommands(input string) {
	cmd, args, _ := strings.Cut(input, " ")

	cmd = strings.ToLower(cmd)
	command := commandList[cmd]

	if command != nil {
		command.goDo(play, args)
	} else {
		cmdListCmds(play.desc)
	}
}

func (play *playerData) send(format string, args ...any) {
	if play.desc == nil {
		return
	}
	play.desc.sendln(format, args...)
}

func (play *playerData) sendToPlaying(format string, args ...any) {
	for _, target := range descList {
		if target.state == CON_PLAYING {
			target.sendln(format, args...)
		}
	}
}

func cmdListCmds(desc *descData) {
	desc.sendln("\r\nCommands:\r\n%v", strings.Join(cmdList, "\r\n"))
}

func (play *playerData) quit(doClose bool) {
	play.desc.sendln(textFiles["aurevoir"])

	if doClose {
		play.desc.state = CON_DISCONNECTED
	} else {
		play.desc.state = CON_CHAR_LIST
		gCharList(play.desc)

		play.desc.inputLock.Lock()
		play.desc.lineBuffer = []string{}
		play.desc.numLines = 0
		play.desc.inputLock.Unlock()
	}
}

func (desc *descData) enterWorld() {
	desc.state = CON_NEWS
	desc.player.sendToPlaying("%v has arrived.", desc.account.tempCharName)

}
