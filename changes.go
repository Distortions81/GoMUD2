package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

type ChangeListData struct {
	Version  int
	dirty    bool
	Changes  []*changeData
	Modified time.Time
}

type changeData struct {
	DateText string
	Text     string

	AddedBy  string
	Modified time.Time
}

var changeList ChangeListData

func saveChanges() {

	if !changeList.dirty {
		return
	}

	tempList := changeList
	changeList.dirty = false

	go func(tempList ChangeListData) {
		tempList.Version = CHANGES_VERSION
		filePath := DATA_DIR + CHANGES_FILE

		outbuf := new(bytes.Buffer)
		enc := json.NewEncoder(outbuf)
		enc.SetIndent("", "\t")

		err := enc.Encode(&tempList)
		if err != nil {
			critLog("saveChanges: enc.Encode: %v", err.Error())
			return
		}

		err = saveFile(filePath, outbuf.Bytes())
		if err != nil {
			critLog("saveChanges: saveFile failed %v", err.Error())
			return
		}
	}(tempList)
}

func unreadChanges(player *characterData) int {
	var count int
	for _, item := range changeList.Changes {
		if item.Modified.IsZero() {
			continue
		}
		if item.Modified.Sub(player.LastChange) > 0 {
			count++
		}
	}
	return count
}

func cmdChanges(player *characterData, input string) {
	if input == "" || strings.EqualFold(input, "next") {
		if unreadChanges(player) == 0 {
			count := unreadChanges(player)
			if count == 0 {
				player.send("No new unread changes.")
				return
			}
		}
		for _, item := range changeList.Changes {
			if item.Modified.IsZero() {
				continue
			}
			if item.Modified.Sub(player.LastChange) < 0 {
				continue
			}
			player.sendWW("%v: (%v)\r\n%v\r\n", item.DateText, item.AddedBy, item.Text)
			player.LastChange = item.Modified
			break
		}
		count := unreadChanges(player)
		if count > 0 {
			player.send("There are %v more unread changes.", count)
		}
		return
	}

	args := strings.SplitN(input, " ", 2)
	numArgs := len(args)

	if strings.EqualFold(input, "list") {
		for i, item := range changeList.Changes {
			if item.Modified.IsZero() {
				continue
			}
			player.send("#%4v -- %v (%v)", i+1, item.DateText, item.AddedBy)
		}
	} else if strings.EqualFold(input, "check") {
		count := unreadChanges(player)
		if count == 0 {
			player.send("No new unread changes.")
		} else {
			player.send("There are %v unread changes", count)
		}
		return
	} else if player.Level >= LEVEL_IMPLEMENTER {
		if strings.EqualFold(args[0], "add") {
			newChange := &changeData{AddedBy: player.Name}
			changeList.Changes = append(changeList.Changes, newChange)
			player.send("New change created, changes date <text> to set date text.")
			player.curChange = newChange
		} else if strings.EqualFold(args[0], "date") {
			if player.curChange != nil {
				if numArgs == 2 && args[1] != "" {
					player.curChange.DateText = args[1]
					player.send("Date text is now: %v", args[1])
					player.send("To set text: changes text <text>")
				}
			}
		} else if strings.EqualFold(args[0], "text") {
			if numArgs == 2 && args[1] != "" {
				player.curChange.Text = args[1]
				player.send("Text is now: %v", args[1])
				player.send("To save the change, type changes done")
			}
		} else if strings.EqualFold(args[0], "done") {
			player.send("Change closed and saved.")
			player.curChange.Modified = time.Now().UTC()
			changeList.Modified = time.Now().UTC()
			player.curChange = nil
			changeList.dirty = true
		} else {
			player.send("That isn't a valid option.")
		}
	}
}
