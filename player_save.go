package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

func saveCharacters() {
	for _, target := range charList {
		if target.dirty {
			target.saveCharacter()
		}
	}
}

func characterNameAvailable(name string) bool {
	var accs, chars int
	//defer func() { errLog("characterNameAvailable: searched %v accounts and %v characters.", accs, chars) }()

	for _, item := range accountIndex {
		accs++
		desc := descData{}
		desc.loadAccount(item.UUID)
		for _, item := range desc.account.Characters {
			chars++
			if strings.EqualFold(item.Login, name) {
				return false
			}
		}
	}

	return true
}

var charSaveLock sync.Mutex

// Returns true on save
func (player *characterData) saveCharacter() bool {

	charSaveLock.Lock()
	defer charSaveLock.Unlock()

	if player.desc == nil {
		critLog("savePlayer: Nil desc: %v", player.Name)
		return false
	} else if player.desc.account == nil {
		critLog("savePlayer: Nil account: %v", player.Name)
		return false
	} else if player == nil {
		critLog("savePlayer: Nil player.")
		return false
	} else if player.UUID == "" {
		critLog("savePlayer: Player '%v' doesn't have a UUID.", player.Name)
		return false
	}

	player.Version = CHARACTER_VERSION
	player.SaveTime = time.Now().UTC()
	target := *player
	player.dirty = false

	go func(target characterData) {
		fileName := DATA_DIR + ACCOUNT_DIR + target.desc.account.UUID + "/" + target.UUID + ".json"
		outbuf := new(bytes.Buffer)
		enc := json.NewEncoder(outbuf)
		enc.SetIndent("", "\t")

		err := enc.Encode(&target)
		if err != nil {
			critLog("savePlayer: enc.Encode: %v", err.Error())
			return
		}

		err = saveFile(fileName, outbuf.Bytes())
		if err != nil {
			critLog("savePlayer: saveFile failed %v", err.Error())
			return
		}
	}(target)

	return true
}

func (desc *descData) loadCharacter(plrStr string) *characterData {
	if desc == nil || desc.account == nil {
		return nil
	}

	uuid := ""
	for _, target := range desc.account.Characters {
		if target.Login == plrStr {
			uuid = target.UUID
			break
		}
	}
	if uuid == "" {
		errLog("loadPlayer: Player not found in account.")
		return nil
	}

	target := checkPlayingUUID(plrStr, uuid)

	if target != nil {
		target.send(aurevoirBuf)
		target.send("Another connection from your account has forcefully taken over control of this character.")
		target.desc.close()

		desc.character = target
		desc.character.desc = desc
		desc.state = CON_PLAYING

		return target
	} else {
		data, err := readFile(DATA_DIR + ACCOUNT_DIR + desc.account.UUID + "/" + uuid + ".json")
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

func (desc *descData) pLoad(plrStr string) *characterData {

	for _, acc := range accountIndex {
		desc := descData{}
		desc.loadAccount(acc.UUID)
		for _, char := range desc.account.Characters {
			if strings.EqualFold(char.Login, plrStr) {
				data, err := readFile(DATA_DIR + ACCOUNT_DIR + acc.UUID + "/" + char.UUID + ".json")
				if err != nil {
					return nil
				}

				target := &characterData{}
				err = json.Unmarshal(data, target)
				if err != nil {
					errLog("loadPlayer: Unable to unmarshal the data.")
					return nil
				}
				target.desc = &desc
				return target
			}
		}
	}
	return nil
}
