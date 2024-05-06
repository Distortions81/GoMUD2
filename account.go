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

func accountNameAvailable(name string) bool {
	return accountIndex[name] == nil
}

func gCharList(desc *descData) {
	var buf string = "\r\n"
	numChars := len(desc.account.Characters)
	if numChars == 0 {
		desc.sendln("\r\nYou're starting fresh with no characters.")
	} else {
		desc.sendln("\r\nYour characters:")

		for i, item := range desc.account.Characters {
			var playing string
			if target := checkPlayingPrint(item.Login, item.Fingerprint); target != nil {
				playing = " (PLAYING)"
			}
			buf = buf + fmt.Sprintf("#%v: %v%v\r\n", i+1, item.Login, playing)
		}
		buf = buf + "\r\n"
	}
	if numChars < MAX_CHAR_SLOTS {
		buf = buf + "Type 'NEW' to create a new character.\r\n"
	}
	if numChars > 0 {
		buf = buf + "Select a character by #number or name: "
	}
	desc.sendln(buf)
}

func accCreateCharacter(desc *descData) {
	if len(desc.account.Characters) < MAX_CHAR_SLOTS {
		desc.state = CON_CHAR_CREATE
	} else {
		desc.sendln("Character creation limit (%v) reached.\r\nNo new characters can be added.", MAX_CHAR_SLOTS)
	}
}

func gCharSelect(desc *descData, input string) {

	input = strings.TrimSpace(input)
	if strings.EqualFold(input, "new") {
		accCreateCharacter(desc)
		return
	}
	nStr, _ := strings.CutPrefix(input, "#")
	num, err := strconv.Atoi(nStr)
	if err != nil { //Find by name
		for _, item := range desc.account.Characters {
			if !strings.EqualFold(item.Login, input) {
				continue
			}
			if target := checkPlayingPrint(item.Login, item.Fingerprint); target != nil {
				alreadyPlayingWarnVictim(target)
				desc.account.tempString = item.Login
				desc.state = CON_RECONNECT_CONFIRM
				return
			}
			var newPlayer *characterData
			if newPlayer = desc.loadCharacter(item.Login); newPlayer != nil {
				desc.enterWorld(newPlayer)
				return
			} else {
				desc.sendln("Failed to load character %v.")
				critLog("Unable to load characer %v!", item.Login)
				return
			}
		}
		desc.sendln("No matches found for %v.", input)
		return
	} else { //Find by number
		if num > 0 && num <= len(desc.account.Characters) {
			var target *characterData
			if target = checkPlayingPrint(desc.account.Characters[num-1].Login, desc.account.Characters[num-1].Fingerprint); target != nil {
				alreadyPlayingWarnVictim(target)
				desc.account.tempString = desc.account.Characters[num-1].Login
				desc.state = CON_RECONNECT_CONFIRM
				return
			}
			var newPlayer *characterData
			if newPlayer = desc.loadCharacter(desc.account.Characters[num-1].Login); newPlayer != nil {
				desc.enterWorld(newPlayer)
				return
			} else {
				desc.sendln("Error: Unable to load requested character.")
				return
			}
		} else {
			desc.sendln("That character isn't listed.")
		}
	}
}

func gReconnectConfirm(desc *descData, input string) {
	filtered := strings.TrimSpace(input)
	filtered = strings.ToLower(filtered)

	if strings.HasPrefix(filtered, "y") {
		var newPlayer *characterData
		if newPlayer = desc.loadCharacter(desc.account.tempString); newPlayer == nil {
			desc.send("Loading the character failed.")
			desc.close()
			return
		}
	} else {
		desc.state = CON_CHAR_LIST
	}
	desc.account.tempString = ""
}

func gCharNewName(desc *descData, input string) {

	input = nameReduce(input)
	if nameReserved(input) {
		desc.sendln("The name you've chosen for your character is not allowed or is reserved.\r\nPlease try a different name.")
		return
	}

	newNameLen := len(input)
	if newNameLen < MIN_NAME_LEN && newNameLen > MAX_NAME_LEN {
		desc.sendln("Character names must be between %v and %v in length.\r\nPlease choose another.", MIN_NAME_LEN, MAX_NAME_LEN)
		return
	}

	if !characterNameAvailable(input) {
		desc.sendln("Unfortunately, the name you've chosen is already taken.")
		return
	}

	desc.account.tempString = input
	desc.state = CON_CHAR_CREATE_CONFIRM
}

func gCharConfirmName(desc *descData, input string) {
	input = nameReduce(input)
	if input == "" || strings.EqualFold(input, "back") {
		desc.sendln("Okay, we can try a different name.")
		desc.state = CON_CHAR_CREATE
		return
	} else if input == desc.account.tempString {
		if !characterNameAvailable(input) {
			desc.sendln("Unfortunately, the name you've chosen is already taken.")
			return
		}

		desc.character = &characterData{
			Fingerprint: makeFingerprintString(),
			Name:        input,
			desc:        desc,
			valid:       true,
			loginTime:   time.Now(),
		}
		desc.account.Characters = append(desc.account.Characters,
			accountIndexData{Login: desc.account.tempString, Fingerprint: desc.character.Fingerprint, Added: time.Now()})
		desc.character.sendToPlaying("--> {GW{gelcome{x %v to the lands! <--", desc.account.tempString)
		desc.account.ModDate = time.Now()
		desc.account.saveAccount()
		desc.character.saveCharacter()
		desc.enterWorld(desc.character)
	} else {
		desc.sendln("Names did not match. Try again, or type 'back' to choose a new name.")
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

func (acc *accountData) saveAccount() bool {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if acc == nil {
		return true
	} else if acc.Fingerprint == "" {
		critLog("saveAccount: Account '%v' doesn't have a fingerprint.", acc.Login)
		return true
	}
	acc.Version = ACCOUNT_VERSION
	acc.ModDate = time.Now()
	fileName := DATA_DIR + ACCOUNT_DIR + acc.Fingerprint + "/" + ACCOUNT_FILE

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

func alreadyPlayingWarnVictim(target *characterData) {
	target.send(textFiles["warn"])
	target.send("\r\nAnother connection on your account is attempting to play this character.\r\nIf they choose 'yes' to confirm you will be kicked.")
}

func gAlreadyPlayingWarn(desc *descData) {
	desc.send(textFiles["warn"])
	desc.send("\r\nThat character is already playing.\r\nDo you wish to disconnect the other session and take control of the character? (y/N)")
}
