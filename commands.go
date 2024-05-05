package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/hako/durafmt"
)

type commandData struct {
	hint string
	goDo func(player *characterData, data string)
	args []string
}

var commandList = map[string]*commandData{
	"say":    {hint: "sends a message", goDo: cmdSay, args: []string{"message"}},
	"quit":   {hint: "quits and disconnects.", goDo: cmdQuit},
	"logout": {hint: "go back to account character selection.", goDo: cmdLogout},
	"who":    {hint: "show players online", goDo: cmdWho},
	"help":   {hint: "get help", goDo: cmdHelp, args: []string{"command, keyword, name or topic"}},
}

type cmdListItem struct {
	name string
	help string
}

var cmdListStr []cmdListItem

func cmdSay(player *characterData, input string) {

	trimInput := strings.TrimSpace(input)
	player.sendToPlaying("%v: %v", player.desc.character.Name, trimInput)
}

func cmdQuit(player *characterData, input string) {
	player.quit(true)
}

func cmdLogout(player *characterData, input string) {
	player.quit(false)
}

func cmdWho(player *characterData, input string) {
	var buf string = "Players online:\r\n"
	for _, target := range characterList {
		if !target.valid {
			continue
		}
		var idleTime, unlink string
		if time.Since(target.idleTime) > time.Minute {
			idleTime = fmt.Sprintf(" (idle %v)", durafmt.Parse(time.Since(target.idleTime).Round(time.Second)).LimitFirstN(2))
		}
		if target.desc == nil || (target.desc != nil && !target.desc.valid) {
			unlink = " (no link)"
		}
		buf = buf + fmt.Sprintf("%30v -- %v%v%v\r\n", target.Name, durafmt.Parse(time.Since(target.loginTime).Round(time.Second)).LimitFirstN(2), idleTime, unlink)
	}
	player.send(buf)
}

func init() {
	cmdListStr = []cmdListItem{}

	for iName, cmd := range commandList {
		tName := fmt.Sprintf("%15v", iName)
		var buf string
		if cmd.args == nil {
			buf = tName + " -- " + cmd.hint
		} else {
			buf = tName + " -- " + cmd.hint + " : " + iName + " "
			for a, aName := range cmd.args {
				if a > 0 {
					buf = buf + " "
				}
				buf = buf + fmt.Sprintf("<%v>", aName)
			}
		}
		cmdListStr = append(cmdListStr, cmdListItem{name: iName, help: buf})
	}

	//sort.Strings(cmdListStr)
}
