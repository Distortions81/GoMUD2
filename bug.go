package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

const (
	MIN_REPORT  = 4
	MAX_REPORT  = 4096
	MAX_REPORTS = 100
)

var bugList bugListData

type bugListData struct {
	Version int
	Bugs    []bugData
	dirty   bool
}

type reporterData struct {
	Name     string
	UUID     uuidData
	Location LocData
}

type bugData struct {
	Reporter reporterData
	Created  time.Time
	Text     string
	Fixed    bool `json:",omitempty"`
}

func cmdBug(player *characterData, input string) {
	if player.Level >= LEVEL_ADMIN && strings.EqualFold(input, "read") {
		for _, item := range bugList.Bugs {
			player.send("(%v) %v", item.Created, item.Reporter.Name)
			player.send("Message:"+NEWLINE+"%v", item.Text)
			player.send("Fixed: %v", item.Fixed)
		}
		return
	}

	input = strings.TrimSpace(input)
	inputLen := len(input)
	if inputLen <= MIN_REPORT {
		player.send("That is really short... Rejected.")
		return
	} else if inputLen > MAX_REPORT {
		player.send("That seems a bit excessive... Rejected.")
	}
	if player.NumReports > MAX_REPORTS {
		player.send("You already have %v reports recorded. Try our discord instead.")
		return
	}

	player.NumReports++
	target := reporterData{Name: player.Name, UUID: player.UUID, Location: player.Loc}
	report := bugData{Reporter: target, Created: time.Now().UTC(), Text: input}
	bugList.Bugs = append(bugList.Bugs, report)
	bugList.dirty = true
	player.send("Your report has been recorded!")
}

func readBugs() {
	data, err := readFile(DATA_DIR + BUGS_FILE)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &bugList)
	if err != nil {
		critLog("readBugs: Unable to unmarshal the data.")
		return
	}
}

func writeBugs() {
	if !bugList.dirty {
		return
	} else {
		bugList.dirty = false
	}

	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	bugList.Version = BUGS_VERSION

	err := enc.Encode(&bugList)
	if err != nil {
		critLog("writeBugs: enc.Encode: %v", err.Error())
		return
	}

	err = saveFile(DATA_DIR+BUGS_FILE, outbuf.Bytes())
	if err != nil {
		critLog("writeBugs: saveFile failed %v", err.Error())
		return
	}
}
