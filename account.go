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
	numChars := len(desc.account.Characters)

	if numChars <= 0 {
		desc.send("You don't have any characters right now.\r\nType NEW to create one:")
		return
	}
	var buf string = "\r\n"
	for i, item := range desc.account.Characters {
		buf = buf + fmt.Sprintf("#%v: %v\r\n", i+1, item)
	}
	if numChars < MAX_CHAR_SLOTS {
		buf = buf + "Type NEW to create a new character.\r\n"
	}
	buf = buf + "Select a player by number or name: "
	desc.send(buf)
}

func gCharSelect(desc *descData, input string) {
	numChars := len(desc.account.Characters)

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
			for _, item := range desc.account.Characters {
				if strings.EqualFold(item, input) {
					desc.send("DEBUG: Would have loaded: %v", input)

					newPlayer := &playerData{Name: input, desc: desc, valid: true, LoginTime: time.Now()}
					desc.player = newPlayer
					playList = append(playList, newPlayer)
					desc.state = CON_NEWS
					desc.player.sendToPlaying("%v has arrived.", desc.account.tempCharName)
					return
				}
			}
			desc.send("Didn't find a character by the name: %v", input)
		} else {
			if num > 0 && num <= numChars {
				selectedChar := desc.account.Characters[num-1]

				newPlayer := &playerData{Name: selectedChar, desc: desc, valid: true, LoginTime: time.Now()}
				desc.player = newPlayer
				playList = append(playList, newPlayer)
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
	if nameBad(input) {
		desc.send("Sorry, that name is not appropriate.")
		return
	}

	newNameLen := len(input)
	if newNameLen < MIN_NAME_LEN && newNameLen > MAX_NAME_LEN {
		desc.send("Sorry, the name must be more than %v and less than %v. Try again!", MIN_NAME_LEN, MAX_NAME_LEN)
		return
	}

	desc.account.tempCharName = input
	desc.state = CON_CHAR_CREATE_CONFIRM
}

func gCharConfirmName(desc *descData, input string) {
	if input == "" {
		desc.send("Okay, we can try again.")
		desc.state = CON_CHAR_CREATE
		return
	} else if input == desc.account.tempCharName {
		desc.send("Okay, you will be called %v.", input)
		desc.account.Characters = append(desc.account.Characters, desc.account.tempCharName)

		err := saveAccount(desc.account)
		if err {
			desc.send("Unable to save account! Please let moderators know!")
			desc.close()
		} else {
			desc.send("Account created and saved.")
		}

		desc.player = &playerData{Name: input, desc: desc, valid: true, LoginTime: time.Now()}
		desc.player.sendToPlaying("Welcome %v to the lands!", desc.account.tempCharName)

		desc.state = CON_NEWS
	} else {
		desc.send("Names did not match. Try again, or blank line to choose a new name.")
	}
}

func createAccountDir(acc *accountData) error {
	if acc.Fingerprint == "" {
		critLog("account has no fingerprint: %v", acc.Login)
		return fmt.Errorf("no fingerprint")
	}

	err := os.Mkdir(DATA_DIR+ACCOUNT_DIR+acc.Fingerprint, 0755)
	if err != nil {
		critLog("unable to make directory for account: %v", acc.Fingerprint)
		return err
	}
	return nil
}

func saveAccount(acc *accountData) (error bool) {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if acc == nil {
		return true
	} else if acc.Fingerprint == "" {
		critLog("Account '%v' doesn't have a fingerprint.", acc.Login)
		return
	}
	acc.Version = ACCOUNT_VERSION
	fileName := DATA_DIR + ACCOUNT_DIR + acc.Fingerprint + "/" + ACCOUNT_FILE

	acc.ModDate = time.Now()

	err := enc.Encode(&acc)
	if err != nil {
		critLog("saveAccount: enc.Encode: %v", err.Error())
		return true
	}

	err = saveFile(fileName, outbuf.Bytes())
	if err != nil {
		critLog("saveAccount: saveFile failed %v", err.Error())
		return true
	}
	acc.dirty = false
	return false
}

func loadAccount(desc *descData, acc *accountData) error {
	data, err := readFile("file")
	if err != nil {
		return err
	}

	accData := &accountData{}
	err = json.Unmarshal(data, accData)
	if err != nil {
		errLog("loadAccount: Unable to unmarshal the data.")
		return err
	}

	desc.account = accData
	return nil
}

func loadPlayerInex() error {
	data, err := readFile(DATA_DIR + PINDEX_FILE)
	if err != nil {
		return err
	}

	newIndex := []playerIndexData{}

	err = json.Unmarshal(data, &newIndex)
	if err != nil {
		errLog("loadPlayerInex: Unable to unmarshal the data.")
		return err
	}

	playerIndex = newIndex
	return nil
}

func savePlayerInex() error {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	err := enc.Encode(&playerIndex)
	if err != nil {
		critLog("saveAccount: enc.Encode: %v", err.Error())
		return err
	}

	err = saveFile(DATA_DIR+PINDEX_FILE, outbuf.Bytes())
	if err != nil {
		critLog("saveAccount: saveFile failed %v", err.Error())
		return err
	}

	return nil
}
