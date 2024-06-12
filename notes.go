package main

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const MAX_DRAFT_NOTES = 5

type noteListData struct {
	Version int
	UUID    uuidData

	Name             string
	File             string
	PostLvl, ReadLvl int
	Notes            []*noteData
	Modified         time.Time

	dirty bool
}

type noteWhoData struct {
	Name string   `json:",omitempty"`
	UUID uuidData `json:",omitempty"`
	//guildid
}

type noteData struct {
	From noteWhoData

	To      []noteWhoData
	CC      []noteWhoData `json:",omitempty"`
	BCC     []noteWhoData `json:",omitempty"`
	Subject string
	Text    string

	Created  time.Time
	Modified time.Time
}

var noteTypes []noteListData

func (who *noteWhoData) isWho(target *characterData) bool {
	if strings.EqualFold(who.Name, "all") {
		return true
	}
	if who.Name == target.Name && who.UUID.sameUUID(target.UUID) {
		return true
	}
	return false
}

func inWhoList(list []noteWhoData, target *characterData) bool {
	for _, item := range list {
		if item.isWho(target) {
			return true
		}
	}

	return false
}

func (note *noteData) isFor(target *characterData) bool {
	if inWhoList(note.To, target) ||
		inWhoList(note.BCC, target) ||
		inWhoList(note.CC, target) {
		return true
	}

	return false
}

func formatWho(list []noteWhoData, target *characterData, bcc bool) string {
	output := ""
	var count int
	for _, item := range list {
		if bcc && !inWhoList(list, target) {
			continue
		}
		if count > 0 {
			output = output + ", "
		}
		if !strings.EqualFold(item.Name, "all") && item.isWho(target) {
			output = output + "(You)"
		} else {
			output = output + item.Name
		}
		count++
	}

	return output
}

func (noteType *noteListData) unread(player *characterData) (*noteData, int) {
	if noteType == nil {
		return nil, 0
	}

	//If note type was modified before or the same as last read...
	//Then we don't need to check all the notes
	if noteType.Modified.Sub(player.NoteRead[noteType.UUID.toString()]) <= 0 {
		return nil, 0
	}

	count := 0
	var oldestNote *noteData
	for _, item := range noteType.Notes {
		if item.Created.Sub(player.NoteRead[noteType.UUID.toString()]) > 0 {
			if !item.isFor(player) {
				continue
			}
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
	if player.Level >= LEVEL_MODERATOR {
		player.send("Mod options: note <type> setting")
	}
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

func draftNote(player *characterData, input string) bool {
	if !player.DraftNotes.Editing {
		return false
	}

	player.send("Note edit mode goes here. Exit to exit.")

	if strings.EqualFold(input, "exit") {
		player.DraftNotes.Editing = false
	}
	return true
}

func cmdNotes(player *characterData, input string) {
	if draftNote(player, input) {
		return
	}
	if len(player.DraftNotes.DraftNotes) > 0 {
		player.send("You currently have %v unfinished draft notes.")
	}
	args := strings.Split(input, " ")
	numArgs := len(args)

	var noteType *noteListData
	var noteTypePos int
	if input == "" || numArgs == 0 {
		noteSyntax(player)
		return
	}
	if player.Level > LEVEL_BUILDER {
		if strings.EqualFold(args[0], "create") {

			typeName := strings.TrimSpace(args[1])
			fileName := txtTo7bit(typeName)
			if len(fileName) < 2 {
				player.send("That note type name is too short: %v", fileName)
				return
			}
			for _, nt := range noteTypes {
				if strings.EqualFold(fileName, nt.File) ||
					strings.EqualFold(typeName, nt.Name) {
					player.send("There is already a note type called that.")
					player.send("Input: File: %v Name: %v", fileName, typeName)
					player.send("Existing note type: File: %v Name: %v", nt.File, nt.Name)
					return
				}
			}
			newType := noteListData{Version: NOTES_VERSION,
				UUID: makeUUID(), File: fileName, Name: args[1], Modified: time.Now().UTC(), dirty: true}
			noteTypes = append(noteTypes, newType)
			player.send("Created new note type: %v", args[1])
			return
		}
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
			player.send("From: %v", note.From.Name)
			player.send("To: %v", formatWho(note.To, player, false))
			if len(note.CC) > 0 {
				player.send("CC: %v", formatWho(note.CC, player, false))
			}
			if len(note.BCC) > 0 {
				player.send("BCC: %v", formatWho(note.BCC, player, true))
			}
			player.send("Subject: %v", note.Subject)
			player.send(NEWLINE + note.Text)

			markRead(player, noteType, note)
		}
		return
	}
	if strings.EqualFold(args[1], "write") {
		newNote := &noteData{From: noteWhoData{Name: player.Name, UUID: player.UUID}, To: []noteWhoData{{Name: "all"}}, Subject: "", Text: "", Created: time.Now().UTC(), Modified: time.Now().UTC()}
		noteTypes[noteTypePos].Modified = time.Now().UTC()
		player.DraftNotes.DraftNotes = append(player.DraftNotes.DraftNotes, newNote)
		player.DraftNotes.Editing = true
		return
	}
	if strings.EqualFold(args[1], "list") {
		if len(noteType.Notes) == 0 {
			player.send("There aren't any %v notes.", noteType.Name)
			return
		}
		pos := 0
		for _, item := range noteType.Notes {
			if player.Level < LEVEL_MODERATOR && !item.isFor(player) {
				continue
			}
			player.send("#%4v: Subject: %v From: %v", pos+1, item.Subject, item.From.Name)
			pos++
		}
		return
	}
	if player.Level < LEVEL_MODERATOR {
		return
	}
	if strings.EqualFold(args[1], "setting") {
		if numArgs > 2 {
			if strings.EqualFold(args[2], "readLevel") {
				if numArgs > 3 {
					val, err := strconv.ParseUint(args[3], 10, 64)
					if err != nil {
						player.send("That isn't a number.")
						return
					}
					noteTypes[noteTypePos].ReadLvl = int(val)
					noteTypes[noteTypePos].dirty = true
					player.send("%v: read level set to %v.", noteType.Name, val)
					return
				}
				player.send("Change the readLevel to what?")

			}
			if strings.EqualFold(args[2], "postLevel") {
				if numArgs > 3 {
					val, err := strconv.ParseUint(args[3], 10, 64)
					if err != nil {
						player.send("That isn't a number.")
						return
					}
					noteTypes[noteTypePos].PostLvl = int(val)
					noteTypes[noteTypePos].dirty = true
					player.send("%v: post level set to %v.", noteType.Name, val)
					return
				}
				player.send("Change the postLevel to what?")
			}
			return
		} else {
			player.send("Settings available: readLevel, postLevel")
		}
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
