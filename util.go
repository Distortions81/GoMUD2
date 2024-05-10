package main

import (
	"os"
	"sync"
)

var saveFileLock sync.Mutex

// Saves as a temp file, then renames when done.
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
