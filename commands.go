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
	name     string
	noShort  bool
	hint     string
	level    int
	goDo     func(player *characterData, data string)
	args     []string
	hide     bool
	forceArg string
}

// command names and shorthands must be lower case
var commandList = map[string]*commandData{
	//Shorthand
	"ne": {level: LEVEL_ANY, noShort: true, goDo: cmdGo, hide: true, forceArg: "northeast"},
	"nw": {level: LEVEL_ANY, noShort: true, goDo: cmdGo, hide: true, forceArg: "northwest"},
	"se": {level: LEVEL_ANY, noShort: true, goDo: cmdGo, hide: true, forceArg: "southeast"},
	"sw": {level: LEVEL_ANY, noShort: true, goDo: cmdGo, hide: true, forceArg: "southwest"},
	"n":  {level: LEVEL_ANY, noShort: true, goDo: cmdGo, hide: true, forceArg: "north"},
	"s":  {level: LEVEL_ANY, noShort: true, goDo: cmdGo, hide: true, forceArg: "south"},
	"e":  {level: LEVEL_ANY, noShort: true, goDo: cmdGo, hide: true, forceArg: "east"},
	"w":  {level: LEVEL_ANY, noShort: true, goDo: cmdGo, hide: true, forceArg: "west"},
	"u":  {level: LEVEL_ANY, noShort: true, goDo: cmdGo, hide: true, forceArg: "up"},
	"d":  {level: LEVEL_ANY, noShort: true, goDo: cmdGo, hide: true, forceArg: "down"},

	//Anyone
	"go":   {level: LEVEL_ANY, hint: "go", goDo: cmdGo, args: []string{"exit name"}},
	"help": {level: LEVEL_ANY, hint: "get help", goDo: cmdHelp, args: []string{"command, keyword, name or topic"}},
	"look": {level: LEVEL_ANY, hint: "look around the room", goDo: cmdLook},
	"quit": {level: LEVEL_ANY, noShort: true, hint: "quit and disconnect", goDo: cmdQuit},
	"who":  {level: LEVEL_ANY, hint: "show players online", goDo: cmdWho},

	//Newbie
	"say":    {level: LEVEL_NEWBIE, hint: "speak out loud", goDo: cmdSay, args: []string{"message"}},
	"telnet": {level: LEVEL_NEWBIE, hint: "telnet options", goDo: cmdTelnet},
	//Player
	"logout":   {level: LEVEL_PLAYER, noShort: true, hint: "quit and go back to character selection menu", goDo: cmdLogout},
	"tell":     {level: LEVEL_PLAYER, hint: "send a private message", args: []string{"target", "message"}, goDo: cmdTell},
	"tells":    {level: LEVEL_PLAYER, hint: "read pending tells", goDo: cmdTells},
	"chat":     {level: LEVEL_NEWBIE, hint: "chat on a channel", goDo: cmdChat},
	"channels": {level: LEVEL_NEWBIE, hint: "turn chat channels on or off", goDo: cmdChannels, args: []string{"channel command"}},
	//Builder
	"olc": {level: LEVEL_BUILDER, hint: "world editor", goDo: cmdOLC},
	//Mod
	"cinfo": {level: LEVEL_MODERATOR, hint: "shows list of connections and characters in the world", goDo: cmdCinfo},
	//Imp
	"pset": {level: LEVEL_IMPLEMENTOR, hint: "set player parameters", goDo: cmdPset},
}

func cmdOLC(player *characterData, input string) {
	interpOLC(player, input)
}

type cmdListItem struct {
	cmd  *commandData
	name string
	help string
}

var cmdListStr []cmdListItem

func cmdTelnet(player *characterData, input string) {
	if player.desc == nil {
		return
	}
	telnet := player.desc.telnet
	buf := "Telnet options:\r\n"
	termType := "Not detected."
	if telnet.termType != "" {
		termType = telnet.termType
	}
	buf = buf + fmt.Sprintf("Mud client: %v\r\nCharset: %v\r\n", termType, telnet.charset)

	if telnet.options == nil {
		return
	}
	if telnet.options.UTF {
		buf = buf + "Supports UTF8\r\n"
	}
	if telnet.options.SUPGA {
		buf = buf + "Supressing GoAhead\r\n"
	}
	if telnet.options.ANSI256 {
		buf = buf + "Supports 256 color mode\r\n"
	}
	if telnet.options.ANSI24 {
		buf = buf + "Supports 24-bit true-color\r\n"
	}
	player.send(buf)
}

func cmdSay(player *characterData, input string) {
	trimInput := strings.TrimSpace(input)
	chatLen := len(trimInput)
	if chatLen == 0 {
		player.send("Say what?")
	} else if chatLen < MAX_CHAT_LENGTH {
		player.sendToRoom("%v says: %v", player.desc.character.Name, trimInput)
		player.send("You say: %v", trimInput)
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
			idleStr := durafmt.Parse(time.Since(target.idleTime).Truncate(time.Minute)).LimitFirstN(2).Format(shortUnits)
			idleStr = strings.ReplaceAll(idleStr, " ", "")
			idleTime = fmt.Sprintf(" (idle %v)", idleStr)
		}
		if target.desc == nil || (target.desc != nil && !target.desc.valid) {
			unlink = " (no link)"
		}
		onlineTime := ""
		if time.Since(target.loginTime) > (time.Minute * 5) {
			onlineTime = durafmt.Parse(time.Since(target.loginTime).Truncate(time.Minute)).LimitFirstN(2).Format(shortUnits)
			onlineTime = strings.ReplaceAll(onlineTime, " ", "")
		}
		buf = buf + fmt.Sprintf("%31v - %v%v%v\r\n", target.Name, onlineTime, idleTime, unlink)
	}
	uptime := durafmt.Parse(time.Since(bootTime).Truncate(time.Second)).LimitFirstN(2).Format(shortUnits)
	uptime = strings.ReplaceAll(uptime, " ", "")
	buf = buf + fmt.Sprintf("\r\n%v players online. Uptime: %v", numPlayers, uptime)
	player.send(buf)
}

func init() {
	shortUnits, _ = durafmt.DefaultUnitsCoder.Decode("y:yrs,wk:wks,d:d,h:h,m:m,s:s,ms:ms,us:us")

	cmdListStr = []cmdListItem{}

	for iName, cmd := range commandList {
		if cmd.hide {
			continue
		}
		cmd.name = iName
		tName := fmt.Sprintf("%10v", iName)
		var buf string
		buf = fmt.Sprintf("%v -- %v :", tName, cmd.hint)
		if cmd.args != nil {
			for _, aName := range cmd.args {
				buf = buf + fmt.Sprintf(" <%v>", aName)
			}
		}
		buf = buf + fmt.Sprintf(" (%v)", levelName[cmd.level])
		cmdListStr = append(cmdListStr, cmdListItem{name: iName, help: buf, cmd: cmd})
	}

	sort.Slice(cmdListStr, func(i, j int) bool {
		if cmdListStr[i].cmd.level == cmdListStr[j].cmd.level {
			return cmdListStr[i].name < cmdListStr[j].name
		} else {
			return cmdListStr[i].cmd.level < cmdListStr[j].cmd.level
		}
	})
}
