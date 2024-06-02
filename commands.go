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
	name       string
	noShort    bool
	hint       string
	level      int
	goDo       func(player *characterData, data string)
	args       []string
	hide       bool
	forceArg   string
	disabled   bool
	olcMode    int
	noAutoHelp bool
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
	"go":        {level: LEVEL_ANY, hint: "go", goDo: cmdGo, args: []string{"exit name"}, noAutoHelp: true},
	"help":      {level: LEVEL_ANY, hint: "get help", goDo: cmdHelp, args: []string{"command, keyword, name or topic"}, noAutoHelp: true},
	"say":       {level: LEVEL_ANY, hint: "speak out loud", goDo: cmdSay, args: []string{"message"}},
	"emote":     {level: LEVEL_ANY, hint: "emote", goDo: cmdEmote, args: []string{"message"}},
	"tell":      {level: LEVEL_ANY, hint: "send a private message", args: []string{"target", "message"}, goDo: cmdTell, noAutoHelp: true},
	"tells":     {level: LEVEL_ANY, hint: "read pending tells", goDo: cmdTells},
	"channels":  {level: LEVEL_ANY, hint: "turn chat channels on or off", goDo: cmdChannels, args: []string{"channel command"}},
	"look":      {level: LEVEL_ANY, hint: "look around the room", goDo: cmdLook, noAutoHelp: true},
	"who":       {level: LEVEL_ANY, hint: "show players online", goDo: cmdWho, noAutoHelp: true},
	"ignore":    {level: LEVEL_ANY, hint: "ignore someone. add 'silent' to silently ignore", goDo: cmdIgnore, args: []string{"player name", "silent"}},
	"config":    {level: LEVEL_ANY, hint: "configure your prefrences", goDo: cmdConfig, args: []string{"1 or more config options to toggle"}, noAutoHelp: true},
	"telnet":    {level: LEVEL_ANY, hint: "telnet options", goDo: cmdTelnet, noAutoHelp: true},
	"quit":      {level: LEVEL_ANY, noShort: true, hint: "quit and disconnect", goDo: cmdQuit, noAutoHelp: true},
	"license":   {level: LEVEL_ANY, noShort: true, hint: "See MUD's version number and license information.", goDo: cmdLicense},
	"note":      {level: LEVEL_ANY, hint: "read notes", goDo: cmdNotes, args: []string{"note type", "list, next"}, noAutoHelp: true},
	"crazytalk": {level: LEVEL_ANY, hint: "global chat with ascii-art text", goDo: cmdCrazyTalk, args: []string{"font", "message"}, hide: true},
	"charlist":  {level: LEVEL_ANY, hint: "see your list of characters", goDo: cmdCharList, noAutoHelp: true},
	"bug":       {level: LEVEL_ANY, hint: "Report a bug or typo in the game", goDo: cmdBug, args: []string{"report message"}, noAutoHelp: true},
	"logout":    {level: LEVEL_ANY, noShort: true, hint: "quit and go back to character selection menu", goDo: cmdLogout, noAutoHelp: true},
	"stats":     {level: LEVEL_ANY, hint: "show some mud stats", goDo: cmdStat, noAutoHelp: true},

	//Builder/mod/imm
	"olc":     {level: LEVEL_BUILDER, hint: "world editor", goDo: cmdOLC, args: []string{"room", "asave", "dig"}},
	"coninfo": {level: LEVEL_MODERATOR, hint: "shows list of connections and characters in the world", goDo: cmdConInfo, noAutoHelp: true},
	"pset":    {level: LEVEL_MODERATOR, hint: "set player parameters", goDo: cmdPset, args: []string{"target", "level", "level-number"}, noAutoHelp: true},
	"disable": {level: LEVEL_MODERATOR, hint: "disable/enable a command or channel", goDo: cmdDisable, args: []string{"command/channel", "name of command or channel"}, noAutoHelp: true},
	"blocked": {level: LEVEL_MODERATOR, hint: "Shows blocked connections", args: []string{"add or delete", "hostname or ip"}, goDo: cmdBlocked},
	"boom":    {level: LEVEL_MODERATOR, hint: "Boom a message at everyone", args: []string{"message"}, goDo: cmdBoom, noAutoHelp: true},
	"ban":     {level: LEVEL_MODERATOR, hint: "Ban a character", args: []string{"target", "reason"}, goDo: cmdBan, noAutoHelp: true},
}

func cmdStat(player *characterData, input string) {
	var ppulse uint64
	var fpulse uint64
	var fpeak, ppeak int64
	for x := 0; x < historyLen-1; x++ {
		fpulse += uint64(fullPulseHistory[x])
		ppulse += uint64(partialPulseHistory[x])
		if fullPulseHistory[x] > int64(fpeak) {
			fpeak = fullPulseHistory[x]
		}
		if partialPulseHistory[x] > int64(ppeak) {
			ppeak = partialPulseHistory[x]
		}
	}
	fpulse = fpulse / uint64(historyLen)
	ppulse = ppulse / uint64(historyLen)

	player.send("\r\nMud load: Pulse: %3.4f%% / Window: %3.2f%%",
		(float64(ppulse)/float64(ROUND_LENGTH_uS))*100.0,
		(float64(fpulse)/float64(ROUND_LENGTH_uS))*100.0)
	player.send("Pulse time: %v (%v peak, %v 5m peak)",
		durafmt.ParseShort(time.Duration(ppulse*uint64(time.Microsecond))),
		durafmt.ParseShort(time.Duration(peakPartialPulse)*time.Microsecond),
		durafmt.ParseShort(time.Duration(ppeak)*time.Microsecond))

	player.send("Window time: %v (%v peak, %v 5m peak)",
		durafmt.ParseShort(time.Duration(fpulse*uint64(time.Microsecond))),
		durafmt.ParseShort(time.Duration(peakFullPulse)*time.Microsecond),
		durafmt.ParseShort(time.Duration(fpeak)*time.Microsecond))
	player.send("(%v averaged)",
		durafmt.ParseShort(time.Duration(time.Duration(historyLen)/(1000000/ROUND_LENGTH_uS)*time.Second)))
}

func cmdLicense(player *characterData, input string) {
	player.send(LICENSE)
}

func cmdSay(player *characterData, input string) {
	if player.Config.hasFlag(CONFIG_DEAF) {
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
	if player.Config.hasFlag(CONFIG_DEAF) {
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
	if player.Config.hasFlag(CONFIG_HIDDEN) {
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
		if target.Config.hasFlag(CONFIG_HIDDEN) {
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
