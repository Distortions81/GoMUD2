package main

import (
	"strings"
	"time"
)

func draftNote(player *characterData, input string) bool {
	if !player.DraftNotes.Editing {
		player.dirty = true
		return false
	}

	if strings.EqualFold(input, "exit") {
		player.DraftNotes.Editing = false
		player.dirty = true
		goForce(player, "")
		return true
	}
	if strings.EqualFold(input, "wordwrap") {
		if player.DraftNotes.Current.WordWrap {
			player.DraftNotes.Current.WordWrap = false
		} else {
			player.DraftNotes.Current.WordWrap = true
		}
		player.send("wordwrap is now %v", boolToText(player.DraftNotes.Current.WordWrap))
		player.dirty = true
	}
	return true
}

func createNewNote(player *characterData, noteTypeID uuidData) {
	newNote := &noteData{From: noteWhoData{Name: player.Name, UUID: player.UUID}, To: []noteWhoData{{Name: "all"}}, Subject: "", Text: "", Created: time.Now().UTC(), Modified: time.Now().UTC()}
	player.DraftNotes.DraftNotes = append(player.DraftNotes.DraftNotes, newNote)
	player.DraftNotes.Current = newNote
	player.DraftNotes.Editing = true
}
