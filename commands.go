package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hako/durafmt"
)

const MAX_CHAT_LENGTH = 2048

var shortUnits durafmt.Units

type commandData struct {
	name    string
	noShort bool
	hint    string
	goDo    func(player *characterData, data string)
	args    []string
}

// command names and shorthands must be lower case
var commandList = map[string]*commandData{
	"say":    {hint: "sends a message", goDo: cmdSay, args: []string{"message"}},
	"quit":   {noShort: true, hint: "quit and disconnect.", goDo: cmdQuit},
	"logout": {noShort: true, hint: "quit and go back to character selection menu.", goDo: cmdLogout},
	"who":    {hint: "show players online", goDo: cmdWho},
	"help":   {hint: "get help", goDo: cmdHelp, args: []string{"command, keyword, name or topic"}},
	"cinfo":  {hint: "show desc & char lists.", goDo: cmdCinfo},
	"look":   {hint: "look around the room.", goDo: cmdLook},
	"go":     {hint: "go", goDo: cmdGo, args: []string{"exit name"}},
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
	chatLen := len(trimInput)
	if chatLen == 0 {
		player.send("Say what?")
	} else if chatLen < MAX_CHAT_LENGTH {
		player.sendToPlaying("%v: %v", player.desc.character.Name, trimInput)
	} else {
		player.send("That is a wall of text. Maybe consider mailing it?")
	}
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
		if time.Since(target.idleTime) >= (time.Minute * 3) {
			idleStr := durafmt.Parse(time.Since(target.idleTime).Round(time.Minute)).LimitFirstN(2).Format(shortUnits)
			idleStr = strings.ReplaceAll(idleStr, " ", "")
			idleTime = fmt.Sprintf(" (idle %v)", idleStr)
		}
		if target.desc == nil || (target.desc != nil && !target.desc.valid) {
			unlink = " (no link)"
		}
		onlineTime := ""
		if time.Since(target.loginTime) > (time.Minute * 5) {
			onlineTime = durafmt.Parse(time.Since(target.loginTime).Round(time.Minute)).LimitFirstN(2).Format(shortUnits)
			onlineTime = strings.ReplaceAll(onlineTime, " ", "")
		}
		buf = buf + fmt.Sprintf("%31v - %v%v%v\r\n", target.Name, onlineTime, idleTime, unlink)
	}
	uptime := durafmt.Parse(time.Since(bootTime).Round(time.Second)).LimitFirstN(2).Format(shortUnits)
	uptime = strings.ReplaceAll(uptime, " ", "")
	buf = buf + fmt.Sprintf("\r\n%v players online. Uptime: %v", numPlayers, uptime)
	player.send(buf)
}

func init() {
	shortUnits, _ = durafmt.DefaultUnitsCoder.Decode("y:yrs,wk:wks,d:d,h:h,m:m,s:s,ms:ms,us:us")

	cmdListStr = []cmdListItem{}

	for iName, cmd := range commandList {
		cmd.name = iName
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
