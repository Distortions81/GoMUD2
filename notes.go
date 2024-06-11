package main

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type noteListData struct {
	Version int
	UUID    uuidData

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

var noteTypes []noteListData

func (noteType *noteListData) unread(player *characterData) (*noteData, int) {
	if noteType == nil {
		return nil, 0
	}

	count := 0
	var oldestNote *noteData
	for _, item := range noteType.Notes {
		if item.Created.Sub(player.NoteRead[noteType.UUID.toString()]) > 0 {
			count++
			if oldestNote == nil || item.Created.Sub(oldestNote.Created) < 0 {
				oldestNote = item
			}
		}
	}
	return oldestNote, count
}

func noteSyntax(player *characterData) {
	listNoteTypes(player)
	player.send("Syntax: note <type> read or list")
}

func markRead(player *characterData, noteType *noteListData, note *noteData) {
	if noteType == nil || note == nil {
		return
	}
	if player.NoteRead[noteType.UUID.toString()].Sub(note.Created) < 0 {
		player.NoteRead[noteType.UUID.toString()] = note.Created
		player.dirty = true
	}
}

func cmdNotes(player *characterData, input string) {
	args := strings.SplitN(input, " ", 2)
	numArgs := len(args)

	var noteType *noteListData
	var noteTypePos int
	if input == "" || numArgs == 0 {
		noteSyntax(player)
		return
	}
	if player.Level > LEVEL_BUILDER && strings.EqualFold(args[0], "create") {
		newType := noteListData{Version: NOTES_VERSION,
			UUID: makeUUID(), File: txtTo7bit(args[1]), Name: args[1], Modified: time.Now().UTC(), dirty: true}
		noteTypes = append(noteTypes, newType)
		player.send("Created new note type: %v", args[1])
		return
	}
	for ntp, item := range noteTypes {
		if strings.EqualFold(item.Name, args[0]) {
			noteType = &item
			noteTypePos = ntp
			break
		}
	}
	if noteType == nil {
		player.send("That isn't a valid note type.")
		listNoteTypes(player)
		return
	}

	if numArgs < 2 {
		noteSyntax(player)
		return
	}

	if strings.EqualFold(args[1], "read") {
		if note, count := noteType.unread(player); count == 0 {
			player.send("No unread %v notes.", noteType.Name)
		} else {
			player.send("On: %v", note.Created.String())
			player.send("From: %v", note.From)
			player.send("To: %v", note.Text)
			player.send("Subject: %v", note.Subject)
			player.send(NEWLINE + note.Text)

			markRead(player, noteType, note)
		}
		return
	}
	if strings.EqualFold(args[1], "write") {
		newNote := &noteData{From: player.Name, To: "all", Subject: "Test", Text: "Blah", Created: time.Now(), Modified: time.Now()}
		noteTypes[noteTypePos].Notes = append(noteTypes[noteTypePos].Notes, newNote)
		noteTypes[noteTypePos].dirty = true
		player.send("%v note created.", noteType.Name)
		return
	}
	if strings.EqualFold(args[1], "list") {
		if len(noteType.Notes) == 0 {
			player.send("There aren't any %v notes.", noteType.Name)
			return
		}
		for i, item := range noteType.Notes {
			player.send("#%4v: Subject: %v From: %v", i+1, item.Subject, item.From)
		}
		return
	}
}

func readNotes() {
	noteLock.Lock()
	defer noteLock.Unlock()

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
			readNoteFile(item.Name())
		}
	}

	//Sort by name
	sort.Slice(noteTypes, func(i, j int) bool {
		return noteTypes[i].Name < noteTypes[j].Name
	})
}

func readNoteFile(fileName string) {
	filePath := DATA_DIR + NOTES_DIR + fileName
	data, err := readFile(filePath)
	if err != nil {
		return
	}

	new := noteListData{File: filePath}
	err = json.Unmarshal(data, &new)
	if err != nil {
		critLog("readNote: Unable to unmarshal note file: %v", filePath)
	}
	noteTypes = append(noteTypes, new)
	errLog("Loaded note type %v", new.Name)
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

		go func(tempList noteListData) {
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

func listNoteTypes(player *characterData) {
	if len(noteTypes) == 0 {
		player.send("Sorry, there aren't any note types available right now.")
	} else {
		player.send("What note type?")
		for _, item := range noteTypes {
			player.send("%v", item.Name)
		}
	}
}
