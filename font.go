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
		player.send(fontList)
		player.send("What font?")
		return
	}

	if len(args[1]) > MAX_CRAZY_INPUT {
		player.send("That message is too long.")
		return
	}

	asciiMsg, err := figletlib.TXTToAscii(args[1], args[0], "left", 0)
	if err != nil {
		player.send("Sorry, that isn't a valid font.")
		return
	}

	if len(asciiMsg) > MAX_CRAZY_OUTPUT {
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

var fontList string

func updateFontList() error {
	dir, err := os.ReadDir(figletlib.FONT_DIR)
	if err != nil {
		critLog("Unable to read font directory: %v -- %v", figletlib.FONT_DIR, err.Error())
		return err
	}
	for _, item := range dir {
		if item.IsDir() {
			continue
		}
		if strings.HasSuffix(item.Name(), ".flf") {
			fontList = fontList + strings.TrimSuffix(item.Name(), ".flf") + ", "
			fontList = wordWrap(fontList, 80)
		}
	}

	return nil
}
