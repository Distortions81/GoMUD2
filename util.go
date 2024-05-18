package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var saveFileLock sync.Mutex

// Saves as a temp file, then renames when done.
func saveFile(filePath string, data []byte) error {
	saveFileLock.Lock()
	defer saveFileLock.Unlock()

	//Save as unique temp file
	tBuf := fmt.Sprintf("%v", time.Now().UTC().UnixNano())
	tmpName := filePath + "-" + tBuf + ".tmp"
	err := os.WriteFile(tmpName, data, 0755)
	if err != nil {
		critLog("saveFile: ERROR: failed to write file: %v", err.Error())
	}

	//Rename to requested filename
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

func (f Bitmask) HasFlag(flag Bitmask) bool { return f&flag != 0 }
func (f *Bitmask) AddFlag(flag Bitmask)     { *f |= flag }
func (f *Bitmask) ClearFlag(flag Bitmask)   { *f &= ^flag }
func (f *Bitmask) ToggleFlag(flag Bitmask)  { *f ^= flag }
