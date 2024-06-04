package main

import (
	"bytes"
	"encoding/json"
	"strings"
)

type serverSettingsData struct {
	Version          int
	NewLock, ModOnly bool
}

var servSet serverSettingsData

func cmdServSet(player *characterData, input string) {

	if input == "" {
		player.send("NewLock: %v, ModOnly: %v",
			boolToText(servSet.ModOnly), boolToText(servSet.NewLock))
		player.send("settings: <option name> toggles specified setting on/off")
		return
	}

	if strings.EqualFold(input, "modonly") {
		if servSet.ModOnly {
			servSet.ModOnly = false
		} else {
			servSet.ModOnly = true
		}
		servSet.saveSettings()
		cmdServSet(player, "")
		player.send("ModeratorOnly is now %v.", boolToText(servSet.ModOnly))

	} else if strings.EqualFold(input, "newlock") {
		if servSet.NewLock {
			servSet.NewLock = false
		} else {
			servSet.NewLock = true
		}
		servSet.saveSettings()
		cmdServSet(player, "")
		player.send("NewLock is now %v.", boolToText(servSet.NewLock))

	} else {
		player.send("That isn't a valid option.")
	}
}
func (set *serverSettingsData) saveSettings() {
	fileName := DATA_DIR + SETTINGS_FILE
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	set.Version = SETTINGS_VERSION
	err := enc.Encode(&set)
	if err != nil {
		critLog("saveSettings: enc.Encode: %v", err.Error())
		return
	}

	err = saveFile(fileName, outbuf.Bytes())
	if err != nil {
		critLog("saveSettings: saveFile failed %v", err.Error())
		return
	}
}

func loadSettings() serverSettingsData {
	set := serverSettingsData{}

	data, err := readFile(DATA_DIR + SETTINGS_FILE)
	if err != nil {
		return set
	}

	err = json.Unmarshal(data, &set)
	if err != nil {
		critLog("loadPlayer: Unable to unmarshal the data.")
		return set
	}
	return set
}
