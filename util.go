package main

import (
	"fmt"
	"os"
	"strings"
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
		critLog("readFile Unable to load file: %v", filePath)
		return nil, err
	}
	return data, nil
}
func (player *characterData) sendWW(format string, args ...any) {
	//Format string if args supplied
	var data string
	if args != nil {
		data = fmt.Sprintf(format, args...)
	} else {
		data = format
	}

	if player.Config.hasFlag(CONFIG_NOWRAP) {
		player.send(format, args)
		return
	} else if player.Columns != 0 {
		player.send(wordWrap(data, player.Columns))
		return

	} else if player.desc != nil {
		if player.desc.telnet.Options != nil {
			if player.desc.telnet.Options.Columns != 0 {
				player.send(wordWrap(data, player.desc.telnet.Options.Columns))
				return
			}
		}
	}

	player.send(wordWrap(data, 80))

}

func wordWrap(input string, cols int) string {
	words := strings.Split(input, " ")

	buf := ""
	lineLen := 0
	for i, item := range words {
		if strings.ContainsAny(item, "\n") ||
			strings.ContainsAny(item, "\r") {
			lineLen = 0
		}
		itemLen := len(string(ColorRemove([]byte(item))))
		if lineLen+itemLen > cols {
			buf = buf + "\r\n"
			lineLen = 0
		}
		if i > 0 {
			buf = buf + " " + item
			lineLen++
		} else {
			buf = buf + item
		}
		lineLen += itemLen
	}
	return buf
}
