package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

func characterNameAvailable(name string) bool {
	var accs, chars int
	defer func() { errLog("characterNameAvailable: searched %v accounts and %v characters.", accs, chars) }()

	for _, item := range accountIndex {
		accs++
		desc := descData{}
		desc.loadAccount(item.Fingerprint)
		for _, item := range desc.account.Characters {
			chars++
			if strings.EqualFold(item.Login, name) {
				return false
			}
		}
	}

	return true
}

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

	target := checkPlaying(plrStr, playFingerprint)

	if target != nil {
		target.send(aurevoirBuf)
		target.send("Another connection from your account has forcely taken over control of this character.")
		target.desc.close()

		desc.character = target
		target.desc.character.desc = desc
		desc.state = CON_PLAYING

		return target
	} else {
		data, err := readFile(DATA_DIR + ACCOUNT_DIR + desc.account.Fingerprint + "/" + playFingerprint + ".json")
		if err != nil {
			return nil
		}

		player := &characterData{}
		err = json.Unmarshal(data, player)
		if err != nil {
			errLog("loadPlayer: Unable to unmarshal the data.")
			return nil
		}
		return player
	}
}