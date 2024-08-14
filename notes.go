package main

import (
	"strings"
	"time"
)

const MAX_DRAFT_NOTES = 5

const noteDraftPrompt = "<note edit: view, add, edit <line>, delete <line>, wordwrap, exit>:"
const noteDraftPromptMod = "<note edit: view, add, edit <line>, delete <line>, wordwrap, create, setting exit>:"

func cmdNotes(player *characterData, input string) {
	numDraft := len(player.DraftNotes.DraftNotes)
	if numDraft > 0 {
		player.send("You currently have %v unfinished draft notes.", numDraft)
	}
	args := strings.Split(input, " ")
	numArgs := len(args)

	var noteType *noteTypeData
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
			newType := noteTypeData{Version: NOTES_VERSION,
				UUID: makeUUID(), File: fileName, Name: args[1], Modified: time.Now().UTC(), dirty: true}
			noteTypes = append(noteTypes, newType)
			player.send("Created new note type: %v", args[1])
			return
		}
	}

	var noteTypePos int
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
		readNotes(player, noteType)
		return
	}
	if strings.EqualFold(args[1], "write") {
		createNewNote(player, noteType.UUID)
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
	noteTypeSettings(player, input, noteTypePos)
}
