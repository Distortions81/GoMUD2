package main

import (
	"os"
	"strings"
	"sync"
)

// Saves as a temp file, then renames
var saveFileLock sync.Mutex

func saveFile(filePath string, data []byte) error {
	saveFileLock.Lock()
	defer saveFileLock.Unlock()

	tmpName := filePath + ".tmp"
	err := os.WriteFile(tmpName, data, 0755)
	if err != nil {
		critLog("saveFile: ERROR: failed to write file: %v", err.Error())
	}
	err = os.Rename(tmpName, filePath)
	if err != nil {
		critLog("saveFile: ERROR: failed to rename file: %v", err.Error())
	}
	return err
}

func readFile(filePath string) ([]byte, error) {
	saveFileLock.Lock()
	defer saveFileLock.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		errLog("readFile Unable to load file: %v", filePath)
		return nil, err
	}
	return data, nil
}

// Returns false if name is prohibited
func nameReserved(name string) bool {
	for _, item := range reservedNames {
		if item == name {
			return true
		}
	}

	return false
}

func nameReduce(name string) string {
	name = strings.ToLower(name)
	var newName string
	for _, l := range name {
		if l >= 'a' && l <= 'z' {
			newName = newName + string(l)
		}
	}
	return toTitle(newName)
}

func toTitle(s string) string {
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
