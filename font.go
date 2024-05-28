package main

import (
	"fmt"
	"goMUD2/figletlib"
	"os"
	"strings"
)

const (
	MAX_CRAZY_INPUT  = 128
	MAX_CRAZY_OUTPUT = 3200
)

func cmdCrazyTalk(player *characterData, input string) {
	if player.Channels.hasFlag(1 << CHAT_CRAZY) {
		player.send("You have the crazytalk channel turned off.")
		return
	}
	args := strings.SplitN(input, " ", 2)
	numArgs := len(args)

	if numArgs < 2 {
		player.send(fontListText)
		player.send("What font?")
		return
	}

	if len(args[1]) > MAX_CRAZY_INPUT {
		player.send("That message is too long.")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			player.send("Sorry, something went wrong rendering that font.")
			return
		}
	}()

	lowerArg := strings.ToLower(args[0])
	asciiMsg, err := figletlib.TXTToAscii(args[1], fontList[lowerArg], "left", 0)
	if err != nil {
		player.send("Sorry, that isn't a valid font.")
		return
	}

	asciiLen := len(asciiMsg)
	if asciiLen < 1 {
		player.send("Sorry, something went wrong rendering that font.")
	}
	if asciiLen > MAX_CRAZY_OUTPUT {
		player.send("That message is too long.")
		return
	}

	buf := fmt.Sprintf("[CRAZY TALK] %v says:\r\n%v", player.Name, asciiMsg)
	for _, target := range charList {
		if !target.Channels.hasFlag(1 << CHAT_CRAZY) {
			target.send(buf)
		}
	}
}

var fontList map[string]string
var fontListText string

func updateFontList() error {
	dir, err := os.ReadDir(figletlib.FONT_DIR)
	if err != nil {
		critLog("Unable to read font directory: %v -- %v", figletlib.FONT_DIR, err.Error())
		return err
	}
	fontList = make(map[string]string)
	fontListText = ""
	itemNum := 0
	for _, item := range dir {
		if item.IsDir() {
			continue
		}
		name := item.Name()
		if strings.HasSuffix(name, ".flf") {
			fname := strings.TrimSuffix(name, ".flf")
			fname = strings.ToLower(fname)
			fname = strings.ReplaceAll(fname, " ", "")
			fname = strings.ReplaceAll(fname, "-", "")
			fname = strings.ReplaceAll(fname, "_", "")
			fontList[fname] = name

			if itemNum > 0 {
				fontListText = fontListText + ", "
			}
			itemNum++
			fontListText = fontListText + fname
		}
		fontListText = wordWrap(fontListText, 80)
	}

	return nil
}
