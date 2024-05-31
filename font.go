package main

import (
	"goMUD2/figletlib"
	"os"
	"strings"
)

const (
	MAX_CRAZY_INPUT  = 128
	MAX_CRAZY_OUTPUT = 6000
)

func cmdCrazyTalk(player *characterData, input string) {
	args := strings.SplitN(input, " ", 2)
	numArgs := len(args)

	if numArgs < 2 {
		player.sendWW(fontListText)
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

	sendToChannel(player, input, CHAT_CRAZY)
}

func handleCrazy(player *characterData, input string) string {
	width := 80
	if player.desc != nil &&
		player.desc.telnet.Options != nil &&
		!player.Config.hasFlag(CONFIG_NOWRAP) {
		width = player.desc.telnet.Options.TermWidth
	}

	args := strings.SplitN(input, " ", 2)
	numArgs := len(args)

	if numArgs < 2 {
		return "Error."
	}

	asciiMsg, _ := figletlib.TXTToAscii(args[1], args[0], "left", width)
	return asciiMsg
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
	}

	return nil
}
