package main

func init() {
	/* Append global olc commands to mode commands lists */
	for m, mode := range olcModes {
		if mode.list == nil {
			continue
		}
		olcModes[m].list = append(olcModes[m].list, gOLCcmds...)
	}

	for i, item := range olcModes {
		modeToText[i] = item.name
	}
}

func limitUndo(player *characterData) {
	numUndo := len(player.OLCEditor.Undo)

	if numUndo > UNDO_MAX {
		player.OLCEditor.Undo = player.OLCEditor.Undo[:UNDO_MAX]
	}
}
