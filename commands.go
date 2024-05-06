package main

import (
	"fmt"
	"sort"
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
	"cinfo":  {hint: "shows desc and char lists.", goDo: cmdCinfo},
	"look":   {hint: "look around the room.", goDo: cmdLook},
}

type cmdListItem struct {
	name string
	help string
}

var cmdListStr []cmdListItem

func cmdCinfo(player *characterData, input string) {
	player.send("Characters:")
	for _, item := range charList {
		if item.desc != nil {
			player.send("valid: %v: name: %v id: %v", item.valid, item.Name, item.desc.id)
		} else {
			player.send("valid: %v: name: %v (no link)", item.valid, item.Name)
		}
	}
	player.send("\r\nDescriptors:")
	for _, item := range descList {
		player.send("id: %v, addr: %v, state: %v", item.id, item.cAddr, item.state)
	}
}

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
	var tmpCharList []*characterData = charList

	numPlayers := len(tmpCharList)
	if numPlayers > 1 {
		sort.Slice(tmpCharList, func(i, j int) bool {
			if tmpCharList[i].desc == nil {
				return false
			} else if tmpCharList[j].desc == nil {
				return true
			}
			return tmpCharList[i].desc.id < tmpCharList[j].desc.id
		})
	}
	for _, target := range tmpCharList {
		var idleTime, unlink string
		if time.Since(target.idleTime) > time.Minute {
			idleTime = fmt.Sprintf(" (idle %v)", durafmt.Parse(time.Since(target.idleTime).Round(time.Second)).LimitFirstN(2))
		}
		if target.desc == nil || (target.desc != nil && !target.desc.valid) {
			unlink = " (no link)"
		}
		buf = buf + fmt.Sprintf("%30v -- %v%v%v\r\n", target.Name, durafmt.Parse(time.Since(target.loginTime).Round(time.Second)).LimitFirstN(2), idleTime, unlink)
	}
	buf = buf + fmt.Sprintf("\r\n%v players online. Uptime: %v", numPlayers, durafmt.Parse(time.Since(bootTime).Round(time.Second)).LimitFirstN(2))
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

	sort.Slice(cmdListStr, func(i, j int) bool {
		return cmdListStr[i].name < cmdListStr[j].name
	})
}
