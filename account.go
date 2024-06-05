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

const CONNECT_WARN = 3

// Returns true if available
func isAccNameAvail(name string) bool {
	return accountIndex[name] == nil
}

func cmdCharList(player *characterData, input string) {
	if player.desc == nil {
		return
	}

	var buf string = "Characters:\r\n"
	for i, item := range player.desc.account.Characters {
		var playing string
		if target := checkPlayingUUID(item.Login, item.UUID); target != nil {
			if target == player {
				playing = " (THIS IS YOU)"
			} else {
				playing = " (ALSO PLAYING)"
			}
		}
		buf = buf + fmt.Sprintf("#%v: %v%v\r\n", i+1, item.Login, playing)
	}
	player.send(buf)

}

func pCharList(desc *descData) {

	//They logged in, reset the attempt count.
	blockedLock.Lock()
	if blockedMap[desc.ip] != nil && blockedMap[desc.ip].Attempts > CONNECT_WARN {
		critLog("*** Warning ***: account '%v' had %v connection attempts before successfully logging in.", desc.account.Login, blockedMap[desc.ip])
	}
	delete(blockedMap, desc.ip)
	blockedDirty = true
	blockedLock.Unlock()

	var buf string = "\r\n"
	numChars := len(desc.account.Characters)
	if numChars == 0 {
		desc.sendln("\r\nYou currently have no characters!")
	} else {
		desc.sendln("\r\nYour characters:")

		for i, item := range desc.account.Characters {
			var playing string
			if target := checkPlayingUUID(item.Login, item.UUID); target != nil {
				playing = " (PLAYING)"
			}
			buf = buf + fmt.Sprintf("#%v: %v%v\r\n", i+1, item.Login, playing)
		}
		buf = buf + "\r\n"
	}
	buf = buf + "Type 'options' for more options.\r\n"
	if numChars < MAX_CHAR_SLOTS {
		buf = buf + "Type NEW or desired name to create a new character.\r\n"
	}
	if numChars > 0 {
		buf = buf + "Select a character by #number or name: "
	}
	desc.sendln(buf)
}

func gCharSelect(desc *descData, input string) {

	input = strings.TrimSpace(input)
	if strings.EqualFold(input, "new") {
		canMakeCharacter(desc, input)
		return
	}
	if strings.EqualFold(input, "options") {
		desc.state = CON_OPTIONS
		return
	}
	nStr, _ := strings.CutPrefix(input, "#")
	num, err := strconv.Atoi(nStr)
	if err != nil { //Find by name
		for _, item := range desc.account.Characters {
			if !strings.EqualFold(item.Login, input) {
				continue
			}
			loadchar(desc, item.Login, item.UUID)
		}
		if !canMakeCharacter(desc, input) {
			desc.sendln("I don't see a character by that name.")
		}
	} else if len(desc.account.Characters) > num-1 { //Find by number
		loadchar(desc, desc.account.Characters[num-1].Login, desc.account.Characters[num-1].UUID)
	} else {
		desc.sendln("That isn't a valid choice.")
	}
}

func canMakeCharacter(desc *descData, input string) bool {
	if servSet.NewLock {
		desc.sendln("New characters are currently prohibited.")
		return false
	}
	if len(desc.account.Characters) < MAX_CHAR_SLOTS {
		if strings.EqualFold(input, "new") {
			desc.state = CON_CHAR_CREATE
			return true
		}
		gCharNewName(desc, input)
		return true
	} else {
		desc.sendln("Character creation limit (%v) reached.\r\nNo new characters can be added.", MAX_CHAR_SLOTS)
		return false
	}
}

func loadchar(desc *descData, login string, uuid uuidData) {
	if target := checkPlayingUUID(login, uuid); target != nil {
		for _, item := range target.Banned {
			if !item.Revoked {
				desc.sendln("This character is banned.")
				desc.kill()
				return
			}
		}
		if servSet.NewLock && desc.character.Level < LEVEL_BUILDER {
			desc.send("Logins are currently limited to staff only.")
			return
		}
		alreadyPlayingWarnVictim(target)
		desc.account.tempString = login
		desc.state = CON_RECONNECT_CONFIRM
		return
	}

	var newPlayer *characterData
	if newPlayer = desc.loadCharacter(login); newPlayer != nil {

		for _, item := range newPlayer.Banned {
			if !item.Revoked {
				desc.sendln("This character is banned.")
				desc.kill()
				return
			}
		}
		if servSet.NewLock && newPlayer.Level < LEVEL_BUILDER {
			desc.send("Logins are currently limited to staff only.")
			return
		}

		desc.enterWorld(newPlayer)
	} else {
		desc.sendln("Failed to load character %v.", login)
		critLog("Unable to load characer %v!", login)
		desc.kill()
		return
	}
}

func gReconnectConfirm(desc *descData, input string) {
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)

	if strings.HasPrefix(input, "y") {
		var newPlayer *characterData
		if newPlayer = desc.loadCharacter(desc.account.tempString); newPlayer == nil {
			desc.send("Loading the character failed.")
			desc.kill()
			return
		}
		newPlayer.send("Reconnected!")
		newPlayer.sendToRoom("%v reconnected.")
		cmdLook(newPlayer, "")
		desc.character.checkTells()
	} else {
		desc.state = CON_CHAR_LIST
	}
	desc.account.tempString = ""
}

func gCharNewName(desc *descData, input string) {
	input = titleCaseAlphaOnly(input)

	if input == "" || strings.EqualFold(input, "cancel") {
		desc.sendln("New character create canceled.")
		desc.state = CON_CHAR_LIST
		return
	}

	if checkCharacterName(desc, input) {
		desc.account.tempString = input
		desc.state = CON_CHAR_CREATE_CONFIRM
	}
}

func checkCharacterName(desc *descData, input string) bool {
	input = titleCaseAlphaOnly(input)
	if isNameReserved(input) {
		desc.sendln("The name you've chosen for your character is not allowed or is reserved.\r\nPlease try a different name.")
		return false
	}

	newNameLen := len(input)
	if newNameLen < MIN_NAME_LEN && newNameLen > MAX_NAME_LEN {
		desc.sendln("Character names must be between %v and %v in length.\r\nPlease choose another.", MIN_NAME_LEN, MAX_NAME_LEN)
		return false
	}

	if !characterNameAvailable(input) {
		desc.sendln("Unfortunately, the name you've chosen is already taken.")
		return false
	}
	return true
}

func gCharConfirmName(desc *descData, input string) {
	input = titleCaseAlphaOnly(input)
	if input == desc.account.tempString {

		if !checkCharacterName(desc, input) {
			desc.state = CON_CHAR_LIST
			return
		}

		desc.character = &characterData{
			UUID:      makeUUID(),
			Name:      input,
			desc:      desc,
			valid:     true,
			loginTime: time.Now(),
			Loc:       LocData{AreaUUID: sysAreaUUID, RoomUUID: sysRoomUUID},
		}
		desc.account.Characters = append(desc.account.Characters,
			accountIndexData{Login: desc.account.tempString, UUID: desc.character.UUID, Added: time.Now().UTC()})

		desc.account.ModDate = time.Now().UTC()
		if desc.account.saveAccount() {
			if desc.character.saveCharacter() {
				desc.enterWorld(desc.character)
			} else {
				desc.kill()
			}
		}

		desc.character.sendToPlaying("--> {GW{gelcome{x %v to the lands! <--", desc.account.tempString)
		desc.sendln("To see a list of commands type: HELP COMMANDS")
	} else {
		desc.state = CON_CHAR_CREATE
		desc.sendln("Names did not match.")
	}
}

func (acc *accountData) createAccountDir() error {
	if !acc.UUID.hasUUID() {
		critLog("createAccountDir: account has no UUID: %v", acc.Login)
		return fmt.Errorf("no UUID")
	}

	path := DATA_DIR + ACCOUNT_DIR + acc.UUID.toString()
	if _, err := os.Stat(path); err == nil {
		critLog("Account directory already exists, aborting!")
		return err
	}

	err := os.Mkdir(path, 0755)
	if err != nil {
		critLog("createAccountDir: unable to make directory for account: %v", acc.UUID)
		return err
	}
	return nil
}

// returns true on save
func (acc *accountData) saveAccount() bool {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if acc == nil {
		return false
	} else if !acc.UUID.hasUUID() {
		critLog("saveAccount: Account '%v' doesn't have a UUID.", acc.Login)
		return false
	}
	acc.Version = ACCOUNT_VERSION
	acc.ModDate = time.Now().UTC()
	fileName := DATA_DIR + ACCOUNT_DIR + acc.UUID.toString() + "/" + ACCOUNT_FILE

	err := enc.Encode(&acc)
	if err != nil {
		critLog("saveAccount: enc.Encode: %v", err.Error())
		return false
	}

	err = saveFile(fileName, outbuf.Bytes())
	if err != nil {
		critLog("saveAccount: saveFile failed %v", err.Error())
		return false
	}
	acc.dirty = false
	return true
}

func (desc *descData) loadAccount(uuid uuidData) error {
	data, err := readFile(DATA_DIR + ACCOUNT_DIR + uuid.toString() + "/" + ACCOUNT_FILE)
	if err != nil {
		critLog("loadAccount: Unable to load account file: %v", err)
		return err
	}

	accData := &accountData{}
	err = json.Unmarshal(data, accData)
	if err != nil {
		critLog("loadAccount: Unable to unmarshal the data: %v", err)
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
		critLog("loadAccountIndex: Unable to unmarshal the data.")
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
	err = saveFile(file, outbuf.Bytes())
	if err != nil {
		critLog("saveAccountIndex: saveFile failed %v", err.Error())
		return err
	}
	mudLog("Account index saved.")

	return nil
}

func alreadyPlayingWarnVictim(target *characterData) {
	target.send(textFiles["warn"])
	target.send("\r\nAnother connection on your account is attempting to play this character.\r\nIf they choose 'yes' to confirm you will be kicked.")
}

func alreadyPlayingWarnPlayer(desc *descData) {
	desc.send(textFiles["warn"])
	desc.send("\r\nThat character is already playing.\r\nDo you wish to disconnect the other session and take control of the character? (y/N)")
}
