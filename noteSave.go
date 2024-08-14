package main

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"sync"
)

func loadNotes() {
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

	new := noteTypeData{File: filePath}
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

		go func(tempList noteTypeData) {
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
