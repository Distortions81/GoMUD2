package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"
)

type NoteListData struct {
	Version  int
	Name     string
	File     string
	Notes    []*noteData
	Modified time.Time

	dirty bool
}

type noteData struct {
	To      string
	From    string
	Subject string
	Text    string

	Created  time.Time
	Modified time.Time
}

var noteTypes []NoteListData

var noteTypeMap map[string]*NoteListData

func readNotes() {
	noteLock.Lock()
	defer noteLock.Unlock()

	noteTypeMap = make(map[string]*NoteListData)

	contents, err := os.ReadDir(DATA_DIR + NOTES_DIR)
	if err != nil {
		critLog("Unable to read notes directory.")
		return
	}

	for _, item := range contents {
		if item.IsDir() {
			continue
		}
		if strings.HasSuffix(item.Name(), ".json") {
			readNote(item.Name())
		}
	}
}

func readNote(fileName string) {
	filePath := DATA_DIR + NOTES_DIR + fileName
	data, err := readFile(filePath)
	if err != nil {
		return
	}

	new := NoteListData{File: filePath}
	err = json.Unmarshal(data, &new)
	if err != nil {
		critLog("readNote: Unable to unmarshal note file: %v", filePath)
	}
	noteTypes = append(noteTypes, new)
	numTypes := len(noteTypes) - 1
	noteTypeMap[new.File] = &noteTypes[numTypes]
	errLog("Loaded notes %v", new.Name)
}

var noteLock sync.Mutex

func saveNotes(force bool) {
	noteLock.Lock()
	defer noteLock.Unlock()

	for _, item := range noteTypes {
		if !item.dirty && !force {
			continue
		}

		tempList := item
		item.dirty = false

		go func(tempList NoteListData) {
			tempList.Version = NOTES_VERSION
			filePath := DATA_DIR + NOTES_DIR + tempList.File + ".json"

			outbuf := new(bytes.Buffer)
			enc := json.NewEncoder(outbuf)
			enc.SetIndent("", "\t")

			err := enc.Encode(&tempList)
			if err != nil {
				critLog("saveNotes: enc.Encode: %v", err.Error())
				return
			}

			err = saveFile(filePath, outbuf.Bytes())
			if err != nil {
				critLog("saveNotes: saveFile failed %v", err.Error())
				return
			}
		}(tempList)
	}
}

func unreadNotes(player *characterData) int {

	var count int
	for _, noteType := range noteTypes {
		for _, item := range noteType.Notes {
			if item.Modified.IsZero() {
				continue
			}
			if item.Modified.Sub(player.LastChange) > 0 {
				count++
			}
		}
	}
	return count
}

func listNoteTypes(player *characterData) {
	player.send("What note type?")
	for _, item := range noteTypes {
		player.send("%v", item.Name)
	}
}

func cmdNotes(player *characterData, input string) {
	parts := strings.SplitN(input, " ", 2)
	var noteType *NoteListData
	if input == "" {
		listNoteTypes(player)
		return
	} else {
		for _, item := range noteTypes {
			if strings.EqualFold(item.Name, parts[0]) {
				noteType = &item
				break
			}
		}
		if noteType == nil {
			player.send("That isn't a valid note type.")
			listNoteTypes(player)
			return
		}
	}

	if input == "" || strings.EqualFold(input, "next") {
		if unreadNotes(player) == 0 {
			count := unreadNotes(player)
			if count == 0 {
				player.send("No new unread changes.")
				return
			}
		}
		for _, item := range noteType.Notes {
			if item.Modified.IsZero() {
				continue
			}
			if item.Modified.Sub(player.LastChange) < 0 {
				continue
			}
			player.sendWW("%v: (%v)\r\n%v\r\n", item.Subject, item.From, item.Text)
			player.LastChange = item.Modified
			break
		}
		count := unreadNotes(player)
		if count > 0 {
			player.send("There are %v more unread changes.", count)
		}
		return
	}

	args := strings.SplitN(input, " ", 2)
	numArgs := len(args)

	if strings.EqualFold(input, "list") {
		for i, item := range noteType.Notes {
			if item.Modified.IsZero() {
				continue
			}
			player.send("#%4v -- %v (%v)", i+1, item.Subject, item.From)
		}
	} else if strings.EqualFold(input, "check") {
		count := unreadNotes(player)
		if count == 0 {
			player.send("No new unread changes.")
		} else {
			player.send("There are %v unread changes", count)
		}
		return
	} else if player.Level >= LEVEL_IMPLEMENTER {
		if strings.EqualFold(args[0], "add") {
			newChange := &noteData{From: player.Name}
			noteType.Notes = append(noteType.Notes, newChange)
			player.send("New change created, changes date <text> to set date text.")
			player.curChange = newChange
		} else if strings.EqualFold(args[0], "date") {
			if player.curChange != nil {
				if numArgs == 2 && args[1] != "" {
					player.curChange.Subject = args[1]
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
			noteType.Modified = time.Now().UTC()
			player.curChange = nil
			noteType.dirty = true
		} else {
			player.send("That isn't a valid option.")
		}
	}
}
