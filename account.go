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
		desc.sendln("You don't have any characters right now.\r\nType NEW to create one:")
		return
	}
	var buf string = "\r\n"
	for i, item := range desc.account.Characters {
		buf = buf + fmt.Sprintf("#%v: %v\r\n", i+1, item.Login)
	}
	if numChars < MAX_CHAR_SLOTS {
		buf = buf + "Type NEW to create a new character.\r\n"
	}
	buf = buf + "Select a player by number or name: "
	desc.sendln(buf)
}

func gCharSelect(desc *descData, input string) {

	if strings.EqualFold(input, "new") {
		if len(desc.account.Characters) < MAX_CHAR_SLOTS {
			desc.sendln("Okay, creating new character!")
			desc.state = CON_CHAR_CREATE
			return
		} else {
			desc.sendln("Sorry, you have hit the limit.")
			return
		}
	} else {
		num, err := strconv.Atoi(input)
		if err != nil {
			for _, item := range desc.account.Characters {
				if strings.EqualFold(item.Login, input) {
					if desc.loadPlayer(item.Login) {
						desc.enterWorld()
						return
					} else {
						desc.sendln("Unable to load that character.")
						return
					}
				}
			}
			desc.sendln("Didn't find a character by the name: %v", input)
		} else {
			if num > 0 && num <= len(desc.account.Characters) {
				if desc.loadPlayer(desc.account.Characters[num-1].Login) {
					desc.enterWorld()
					return
				} else {
					desc.sendln("Unable to load that character.")
					return
				}
			} else {
				desc.sendln("That player doesn't seem to exist.")
			}
		}
	}
}

func gCharNewName(desc *descData, input string) {
	if nameBad(input) {
		desc.sendln("Sorry, that name is not appropriate.")
		return
	}

	newNameLen := len(input)
	if newNameLen < MIN_NAME_LEN && newNameLen > MAX_NAME_LEN {
		desc.sendln("Sorry, the name must be more than %v and less than %v. Try again!", MIN_NAME_LEN, MAX_NAME_LEN)
		return
	}

	desc.account.tempCharName = input
	desc.state = CON_CHAR_CREATE_CONFIRM
}

func gCharConfirmName(desc *descData, input string) {
	if input == "" {
		desc.sendln("Okay, we can try again.")
		desc.state = CON_CHAR_CREATE
		return
	} else if input == desc.account.tempCharName {
		desc.sendln("Okay, you will be called %v.", input)
		desc.player = &playerData{
			Fingerprint: makeFingerprintString(),
			Name:        input,
			desc:        desc,
			valid:       true,
			LoginTime:   time.Now(),
		}
		desc.account.Characters = append(desc.account.Characters,
			accountIndexData{Login: desc.account.tempCharName, Fingerprint: desc.player.Fingerprint})
		desc.player.sendToPlaying("Welcome %v to the lands!", desc.account.tempCharName)
		desc.account.ModDate = time.Now()
		desc.account.saveAccount()
		desc.player.savePlayer()
		desc.state = CON_NEWS
	} else {
		desc.sendln("Names did not match. Try again, or blank line to choose a new name.")
	}
}

func (acc *accountData) createAccountDir() error {
	if acc.Fingerprint == "" {
		critLog("createAccountDir: account has no fingerprint: %v", acc.Login)
		return fmt.Errorf("no fingerprint")
	}

	err := os.Mkdir(DATA_DIR+ACCOUNT_DIR+acc.Fingerprint, 0755)
	if err != nil {
		critLog("createAccountDir: unable to make directory for account: %v", acc.Fingerprint)
		return err
	}
	return nil
}

func (acc *accountData) saveAccount() (error bool) {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if acc == nil {
		return true
	} else if acc.Fingerprint == "" {
		critLog("saveAccount: Account '%v' doesn't have a fingerprint.", acc.Login)
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

func (desc *descData) loadAccount(fingerprint string) error {
	data, err := readFile(DATA_DIR + ACCOUNT_DIR + fingerprint + "/" + ACCOUNT_FILE)
	if err != nil {
		errLog("loadAccount: Unable to load account file: %v", err)
		return err
	}

	accData := &accountData{}
	err = json.Unmarshal(data, accData)
	if err != nil {
		errLog("loadAccount: Unable to unmarshal the data: %v", err)
		return err
	}

	desc.account = accData
	return nil
}

func loadAccountIndex() error {
	file := DATA_DIR + ACCOUNT_DIR + ACC_INDEX_FILE
	data, err := readFile(file)
	if err != nil {
		return err
	}

	newIndex := make(map[string]*accountIndexData)
	err = json.Unmarshal(data, &newIndex)
	if err != nil {
		errLog("loadAccountIndex: Unable to unmarshal the data.")
		return err
	}

	accountIndex = newIndex
	return nil
}

func saveAccountIndex() error {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	err := enc.Encode(&accountIndex)
	if err != nil {
		critLog("saveAccountIndex: enc.Encode: %v", err.Error())
		return err
	}

	file := DATA_DIR + ACCOUNT_DIR + ACC_INDEX_FILE
	tempFile := file + ".tmp"
	err = saveFile(tempFile, outbuf.Bytes())
	if err != nil {
		critLog("saveAccountIndex: saveFile failed %v", err.Error())
		return err
	}
	err = os.Rename(tempFile, file)
	if err != nil {
		critLog("saveAccountIndex: rename failed %v", err.Error())
		return err
	}
	errLog("Account index saved.")

	return nil
}
