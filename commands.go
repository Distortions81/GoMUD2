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
	disabled bool
}

var cmdList []*commandData

// command names and shorthands must be lower case
// use 'disable: true' to disable a command.
var cmdMap = map[string]*commandData{
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
	"go":      {level: LEVEL_ANY, hint: "go", goDo: cmdGo, args: []string{"exit name"}},
	"help":    {level: LEVEL_ANY, hint: "get help", goDo: cmdHelp, args: []string{"command, keyword, name or topic"}},
	"look":    {level: LEVEL_ANY, hint: "look around the room", goDo: cmdLook},
	"quit":    {level: LEVEL_ANY, noShort: true, hint: "quit and disconnect", goDo: cmdQuit},
	"who":     {level: LEVEL_ANY, hint: "show players online", goDo: cmdWho},
	"tells":   {level: LEVEL_ANY, hint: "read pending tells", goDo: cmdTells},
	"ignore":  {level: LEVEL_ANY, hint: "ignore someone. add 'silent' to silently ignore", goDo: cmdIgnore, args: []string{"player name", "silent"}},
	"changes": {level: LEVEL_ANY, hint: "See list of changes made to the MUD.", goDo: cmdChanges, args: []string{"list, next"}},
	"license": {level: LEVEL_ANY, hint: "See MUD's version number and license information.", goDo: cmdLicense},

	//Newbie
	"say":      {level: LEVEL_NEWBIE, hint: "speak out loud", goDo: cmdSay, args: []string{"message"}},
	"emote":    {level: LEVEL_NEWBIE, hint: "emote", goDo: cmdEmote, args: []string{"message"}},
	"telnet":   {level: LEVEL_NEWBIE, hint: "telnet options", goDo: cmdTelnet},
	"chat":     {level: LEVEL_NEWBIE, hint: "chat on a channel", goDo: cmdChat},
	"channels": {level: LEVEL_NEWBIE, hint: "turn chat channels on or off", goDo: cmdChannels, args: []string{"channel command"}},

	//Player
	"logout":   {level: LEVEL_PLAYER, noShort: true, hint: "quit and go back to character selection menu", goDo: cmdLogout},
	"tell":     {level: LEVEL_PLAYER, hint: "send a private message", args: []string{"target", "message"}, goDo: cmdTell},
	"config":   {level: LEVEL_PLAYER, hint: "configure your prefrences.", goDo: cmdConfig, args: []string{"1 or more config options to toggle"}},
	"charlist": {level: LEVEL_PLAYER, hint: "see your list of characters", goDo: cmdCharList},

	//Builder/mod/imm
	"olc":     {level: LEVEL_BUILDER, hint: "world editor", goDo: cmdOLC, args: []string{"room", "asave", "dig"}},
	"coninfo": {level: LEVEL_MODERATOR, hint: "shows list of connections and characters in the world", goDo: cmdConInfo},
	"pset":    {level: LEVEL_IMPLEMENTER, hint: "set player parameters", goDo: cmdPset, args: []string{"player-name", "level", "level-number"}},
	"disable": {level: LEVEL_ADMIN, hint: "disable/enable a command or channel", goDo: cmdDisable, args: []string{"command/channel", "name of command or channel"}},
	"blocked": {level: LEVEL_ADMIN, hint: "Shows blocked connections", args: []string{"add or delete", "hostname or ip"}, goDo: cmdBlocked},
}

func cmdLicense(player *characterData, input string) {
	player.send(LICENSE)
}

func cmdCharList(player *characterData, input string) {
	if player.desc == nil {
		return
	}

	var buf string = "Characters:\r\n"
	for i, item := range player.desc.account.Characters {
		var playing string
		if target := checkPlayingUUID(item.Login, item.UUID); target != nil {
			if target == player {
				playing = " (THIS IS YOU)"
			} else {
				playing = " (ALSO PLAYING)"
			}
		}
		buf = buf + fmt.Sprintf("#%v: %v%v\r\n", i+1, item.Login, playing)
	}
	player.send(buf)

}

func cmdOLC(player *characterData, input string) {
	interpOLC(player, input)
}

func cmdTelnet(player *characterData, input string) {
	if player.desc == nil {
		return
	}

	telnet := player.desc.telnet
	if input == "" {
		buf := "Telnet options:\r\n"
		termType := "Not detected."
		if telnet.termType != "" {
			termType = telnet.termType
		}
		buf = buf + fmt.Sprintf("Selected character map: %v\r\n\r\n", telnet.Charset)
		buf = buf + fmt.Sprintf("Mud client: %v\r\n", termType)

		if telnet.Options == nil {
			return
		}
		if telnet.Options.suppressGoAhead {
			buf = buf + "Supressing Go-Ahead Signal (SUPGA)\r\n"
		}
		if telnet.Options.ColorDisable {
			buf = buf + "Color disabled.\r\n"
		}
		if telnet.Options.ansi256 {
			buf = buf + "Supports 256 color mode\r\n"
		}
		if telnet.Options.ansi24 {
			buf = buf + "Supports 24-bit true-color\r\n"
		}

		player.send(buf)
		if termType == "" {
			player.send("Options: telnet SAVE, UTF, SUPGA, COLOR or the name of a charmap.")
			player.send("To see a list of available character maps, type 'telnet charmaps'")
			player.send("telnet SAVE will, save these settings to your account.")
		}
		return
	}
	if strings.EqualFold("COLOR", input) {
		if player.desc == nil {
			return
		}
		if telnet.Options.ColorDisable {
			telnet.Options.ColorDisable = false
			player.send("ANSI color is now enabled.")
		} else {
			telnet.Options.ColorDisable = true
			player.send("ANSI color is now disabled.")
		}
		return
	}
	if strings.EqualFold("SAVE", input) {
		if player.desc == nil {
			return
		}
		player.desc.account.TelnetSettings = &player.desc.telnet
		player.desc.account.saveAccount()
		player.send("Your telnet settings have been saved to your account.")
		return
	}
	if strings.EqualFold("UTF", input) {
		if telnet.Options.UTF {
			telnet.Options.UTF = false
			player.send("UTF mode disabled.")
		} else {
			telnet.Options.UTF = true
			player.send("UTF mode enabled.")
		}
		player.send("Character map test:")
		player.sendTestString()
		return
	}
	if strings.EqualFold("supga", input) {
		if telnet.Options.suppressGoAhead {
			telnet.Options.suppressGoAhead = false
			player.send("SUPGA mode disabled.")
		} else {
			telnet.Options.suppressGoAhead = true
			player.send("SUPGA mode enabled.")
		}
		return
	}
	if strings.EqualFold("charmaps", input) {
		player.send("Character map list:")
		var buf string
		var count int
		for cname := range charsetList {
			count++
			buf = buf + fmt.Sprintf("%18v", cname)
			if count%4 == 0 {
				buf = buf + "\r\n"
			}
		}
		player.send(buf)
		player.send("To enable UTF, type 'telnet utf'")
		return
	}
	for cname, cset := range charsetList {
		if strings.EqualFold(input, cname) {
			telnet.Charset = cname
			telnet.charMap = cset
			telnet.Options.UTF = false
			player.send("Your character map has been changed to: %v", cname)
			player.send("Character set test:")
			player.sendTestString()
			return
		}
	}
	player.send("That isn't a valid character map.")
}

func (player *characterData) sendTestString() {
	player.send("Falsches Üben von Xylophonmusik quält jeden größeren Zwerg")
}

func cmdSay(player *characterData, input string) {
	if player.Config.HasFlag(CONFIG_DEAF) {
		player.send("You are currently deaf.")
		return
	}
	trimInput := strings.TrimSpace(input)
	chatLen := len(trimInput)
	if chatLen == 0 {
		player.send("Say what?")
	} else if chatLen < MAX_CHAT_LENGTH {
		player.sendToRoom("%v says: %v", player.Name, trimInput)
		player.send("You say: %v", trimInput)
	} else {
		player.send("That is a wall of text. Maybe consider mailing it?")
	}
}

func cmdEmote(player *characterData, input string) {
	if player.Config.HasFlag(CONFIG_DEAF) {
		player.send("You are currently deaf.")
		return
	}
	trimInput := strings.TrimSpace(input)
	chatLen := len(trimInput)
	if chatLen == 0 {
		player.send("Emote what?")
	} else if chatLen < MAX_CHAT_LENGTH {
		player.sendToRoom("%v %v", player.Name, trimInput)
		player.send("%v %v", player.Name, trimInput)
	} else {
		player.send("That seems a bit excessive for an emote.")
	}
}

func cmdQuit(player *characterData, input string) {
	player.quit(true)
}

func cmdLogout(player *characterData, input string) {
	player.quit(false)
}

func cmdWho(player *characterData, input string) {
	if player.Config.HasFlag(CONFIG_HIDDEN) {
		player.send("You are currently hidden.")
		return
	}
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
	buf = buf + fmt.Sprintf("%31v - %v %v %v %v\r\n", "Player name", "level", "time-online", "(idle time)", "(no link)")
	for _, target := range tmpCharList {
		if target.Config.HasFlag(CONFIG_HIDDEN) {
			continue
		}
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
		buf = buf + fmt.Sprintf("%31v - %v %v%v%v\r\n", target.Name, levelName[target.Level], onlineTime, idleTime, unlink)
	}
	uptime := durafmt.Parse(time.Since(bootTime).Truncate(time.Second)).LimitFirstN(2).Format(shortUnits)
	uptime = strings.ReplaceAll(uptime, " ", "")
	buf = buf + fmt.Sprintf("\r\n%v players online. Uptime: %v", numPlayers, uptime)
	player.send(buf)
}

func init() {
	shortUnits, _ = durafmt.DefaultUnitsCoder.Decode("y:yrs,wk:wks,d:d,h:h,m:m,s:s,ms:ms,us:us")

	for name, item := range cmdMap {
		item.name = name
		cmdList = append(cmdList, item)
	}

	//Sort by level and name
	sort.Slice(cmdList, func(i, j int) bool {
		if cmdList[i].level == cmdList[j].level {
			return cmdList[i].name < cmdList[j].name
		} else {
			return cmdList[i].level < cmdList[j].level
		}
	})
}
