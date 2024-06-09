package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var saveFileLock sync.Mutex

func (old DIR) revDir() DIR {

	switch old {
	case DIR_NORTH:
		return DIR_SOUTH
	case DIR_SOUTH:
		return DIR_NORTH
	case DIR_EAST:
		return DIR_WEST
	case DIR_WEST:
		return DIR_EAST
	case DIR_UP:
		return DIR_DOWN
	case DIR_DOWN:
		return DIR_UP
	case DIR_NORTH_EAST:
		return DIR_SOUTH_WEST
	case DIR_SOUTH_WEST:
		return DIR_NORTH_EAST
	case DIR_NORTH_WEST:
		return DIR_SOUTH_EAST
	case DIR_SOUTH_EAST:
		return DIR_NORTH_WEST
	default:
		return old
	}
}

func percentColor(per float64) string {
	if per <= 25 {
		return "{C"
	} else if per <= 50 {
		return "{G"
	} else if per <= 75 {
		return "{Y"
	} else if per <= 85 {
		return "{M"
	} else {
		return "{R"
	}
}

func cEllip(input string, limit int) string {
	inLen := len(input)
	if inLen > limit {
		return input[:limit-3] + "..."
	}

	return input
}

func cText(input string, limit int) string {
	inLen := len(input)
	if inLen > limit {
		return input[:limit]
	}

	return input
}

func boolToText(value bool) string {
	if value {
		return "{gOn{x"
	} else {
		return "{rOff{x"
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

	words := strings.Split(input, " ")
	width := 80

	if player.desc != nil &&
		player.desc.telnet.Options != nil {

		if player.Config.hasFlag(CONFIG_TERMWIDTH) &&
			player.ConfigVals[CONFIG_TERMWIDTH] != nil {

			width = player.ConfigVals[CONFIG_TERMWIDTH].Value
		} else {
			width = player.desc.telnet.Options.TermWidth
		}
	}

	if player.Config.hasFlag(CONFIG_NOWRAP) || len(input) <= width {
		return input
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
			buf = buf + NEWLINE + item
			lineLen = 0
			continue
		}
		//If adding this word to the line would go over, add a newline and reset line width
		if lineLen+itemLen >= width {
			buf = buf + NEWLINE
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
