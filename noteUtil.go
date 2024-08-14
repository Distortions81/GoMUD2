package main

import (
	"fmt"
	"strconv"
	"strings"
)

func noteTypeSettings(player *characterData, input string, noteTypePos int) {
	noteType := noteTypes[noteTypePos]

	args := strings.Split(input, " ")
	numArgs := len(args)

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

func formatNoteWho(list []noteWhoData, target *characterData, bcc bool) string {
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

func (noteType *noteTypeData) unread(player *characterData) (*noteData, int) {
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

func (player *characterData) checkUnreadNotes() {
	buf := ""

	for _, nt := range noteTypes {
		if _, count := nt.unread(player); count != 0 {
			buf = buf + fmt.Sprintf("%v unread %v"+NEWLINE, count, nt.Name)
		}
	}
	if buf == "" {
		player.send("You have no unread notes.")
	} else {
		player.send("Unread notes: (note <type> read)")
		player.send(buf)
	}
}

func noteSyntax(player *characterData) {
	player.checkUnreadNotes()

	listNoteTypes(player)
	player.send("Syntax: note <type> read or list")
	if player.Level >= LEVEL_MODERATOR {
		player.send("Mod options: note <type> setting")
	}
}

func readNotes(player *characterData, noteType *noteTypeData) {
	if note, count := noteType.unread(player); count == 0 {
		player.send("No unread %v notes.", noteType.Name)
	} else {
		player.send("On: %v", note.Created.String())
		player.send("From: %v", note.From.Name)
		player.send("To: %v", formatNoteWho(note.To, player, false))
		if len(note.CC) > 0 {
			player.send("CC: %v", formatNoteWho(note.CC, player, false))
		}
		if len(note.BCC) > 0 {
			player.send("BCC: %v", formatNoteWho(note.BCC, player, true))
		}
		player.send("Subject: %v", note.Subject)
		player.send(NEWLINE + note.Text)

		markRead(player, noteType, note)
	}
}

func markRead(player *characterData, noteType *noteTypeData, note *noteData) {
	if noteType == nil || note == nil {
		return
	}
	if player.NoteRead[noteType.UUID.toString()].Sub(note.Created) < 0 {
		player.NoteRead[noteType.UUID.toString()] = note.Created
		player.dirty = true
	}
}
