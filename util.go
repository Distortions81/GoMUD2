package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var saveFileLock sync.Mutex

func boolToText(value bool) string {
	if value {
		return "On"
	} else {
		return "Off"
	}
}

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

	player.send(player.wordWrap(data))
}

func (player *characterData) wordWrap(input string) string {
	if player.Config.hasFlag(CONFIG_NOWRAP) || len(input) <= 80 {
		return input
	}

	words := strings.Split(input, " ")
	width := 80

	if player.desc != nil &&
		player.desc.telnet.Options != nil &&
		!player.Config.hasFlag(CONFIG_NOWRAP) {
		width = player.desc.telnet.Options.TermWidth
	}

	buf := ""
	lineLen := 0
	for _, item := range words {
		if strings.ContainsAny(item, "\n") ||
			strings.ContainsAny(item, "\r") {
			lineLen = 0
		}
		itemLen := len([]byte(item))
		//Don't bother to check the whole thing if the item itself is larger than width (deco)
		if itemLen > width {
			buf = buf + "\r\n" + item
			lineLen = 0
			continue
		}
		//If adding this word to the line would go over, add a newline and reset line width
		if lineLen+itemLen >= width {
			buf = buf + "\r\n"
			lineLen = 0
		}
		//If not the first word, add a space before adding it
		if lineLen > 0 {
			buf = buf + " " + item
			lineLen++
		} else {
			//Otherwise just add the text
			buf = buf + item
		}
		//Add item len to line
		lineLen += itemLen
	}
	return buf
}
