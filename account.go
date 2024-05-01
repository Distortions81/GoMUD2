package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func gCharList(desc *descData) {
	numChars := len(desc.account.characters)

	if numChars <= 0 {
		desc.send("You don't have any characters right now.\r\nType NEW to create one:")
		return
	}
	var buf string
	for i, item := range desc.account.characters {
		buf = buf + fmt.Sprintf("#%v: %v\r\n", i+1, item)
	}
	if numChars < MAX_CHAR_SLOTS {
		buf = buf + "Type NEW to create a new character.\r\n"
	}
	buf = buf + "Select a player by number or name: "
	desc.send(buf)
}

func gCharSelect(desc *descData, input string) {
	numChars := len(desc.account.characters)

	if strings.EqualFold(input, "new") {
		if numChars < MAX_CHAR_SLOTS {
			desc.send("Okay, creating new character!")
			desc.state = CON_CHAR_CREATE
			return
		} else {
			desc.send("Sorry, you have hit the limit.")
			return
		}
	} else {
		num, err := strconv.Atoi(input)
		if err != nil {
			for _, item := range desc.account.characters {
				if strings.EqualFold(item, input) {
					desc.send("DEBUG: Would have loaded: %v", input)

					desc.player = &playerData{name: input, desc: desc, valid: true, loginTime: time.Now()}
					playList = append(playList, desc.player)
					desc.state = CON_NEWS
					desc.player.sendToPlaying("%v has arrived.", desc.account.tempCharName)
					return
				}
			}
			desc.send("Didn't find a character by the name: %v", input)
		} else {
			if num > 0 && num <= numChars {
				selectedChar := desc.account.characters[num-1]

				desc.player = &playerData{name: selectedChar, desc: desc, valid: true, loginTime: time.Now()}
				playList = append(playList, desc.player)
				desc.state = CON_NEWS
				desc.send("DEBUG: Would have loaded %v", selectedChar)
				return
			} else {
				desc.send("That player doesn't seem to exist.")
			}
		}
	}
}

func gCharNewName(desc *descData, input string) {
	newNameLen := len(input)
	if newNameLen < MIN_NAME_LEN && newNameLen > MAX_NAME_LEN {
		desc.send("Sorry, the name must be more than %v and less than %v. Try again!", MIN_NAME_LEN, MAX_NAME_LEN)
		return
	}

	desc.account.tempCharName = input
	desc.state = CON_CHAR_CREATE_CONFIRM
}

func gCharConfirmName(desc *descData, input string) {
	if input == desc.account.tempCharName {
		desc.send("Okay, you will be called %v.", input)
		desc.account.characters = append(desc.account.characters, desc.account.tempCharName)
		desc.player = &playerData{name: input, desc: desc, valid: true, loginTime: time.Now()}
		desc.player.sendToPlaying("Welcome %v to the lands!", desc.account.tempCharName)
		desc.state = CON_NEWS
	} else {
		desc.send("Names did not match. Try again, or blank line to choose a new name.")
	}
}

func createAccountDir(acc *accountData) error {
	if acc.fingerprint == "" {
		errLog("account has no fingerprint: %v", acc.login)
		return fmt.Errorf("No fingerprint")
	}

	err := os.Mkdir(DATA_DIR+ACCOUNT_DIR+acc.fingerprint, 0755)
	if err != nil {
		errLog("unable to make directory for account: %v", acc.fingerprint)
		return err
	}
	return nil
}

func saveAccount(acc *accountData, asyncSave bool) (error bool) {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if acc == nil {
		return true
	}
	acc.version = ACCOUNT_VERSION
	fileName := DATA_DIR + ACCOUNT_DIR + acc.fingerprint + "/" + ACCOUNT_FILE

	acc.modDate = time.Now()

	if err := enc.Encode(&acc); err != nil {
		errLog("WritePlayer: enc.Encode", err)
		return true
	}

	_, err := os.Create(fileName)

	if err != nil {
		errLog("WritePlayer: os.Create", err)
		return true
	}

	//Async write
	/*
		if asyncSave {
			go writePlayerFile(outbuf, fileName)
		} else {
			writePlayerFile(outbuf, fileName)
		}*/

	acc.dirty = false
	return true
}
