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
	level   int
	goDo    func(player *characterData, data string)
	args    []string
	modargs []string

	hide       bool
	forceArg   string
	disabled   bool
	noAutoHelp bool
	subType    func(player *characterData, list []*commandData)
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
	"help":      {level: LEVEL_ANY, hint: "get help", goDo: cmdHelp, args: []string{"command, keyword, name or topic"}},
	"say":       {level: LEVEL_ANY, hint: "speak out loud", goDo: cmdSay, args: []string{"message"}, noAutoHelp: true},
	"emote":     {level: LEVEL_ANY, hint: "emote", goDo: cmdEmote, args: []string{"message"}, noAutoHelp: true},
	"tell":      {level: LEVEL_ANY, hint: "send a private message, even if offline", args: []string{"target", "message"}, goDo: cmdTell, noAutoHelp: true},
	"tells":     {level: LEVEL_ANY, hint: "read pending tells", goDo: cmdTells, noAutoHelp: true},
	"channels":  {level: LEVEL_ANY, hint: "turn chat channels on or off", goDo: cmdChannels, args: []string{"channel"}},
	"look":      {level: LEVEL_ANY, hint: "look around the room", goDo: cmdLook, noAutoHelp: true},
	"who":       {level: LEVEL_ANY, hint: "show players online", goDo: cmdWho, noAutoHelp: true},
	"ignore":    {level: LEVEL_ANY, hint: "ignore someone", goDo: cmdIgnore, args: []string{"player name", "silent"}},
	"config":    {level: LEVEL_ANY, hint: "configure your prefrences", goDo: cmdConfig, args: []string{"option"}},
	"telnet":    {level: LEVEL_ANY, hint: "telnet options", args: []string{"UTF", "charmaps"}, goDo: cmdTelnet, noAutoHelp: true},
	"quit":      {level: LEVEL_ANY, noShort: true, hint: "quit and disconnect", goDo: cmdQuit, noAutoHelp: true},
	"license":   {level: LEVEL_ANY, noShort: true, hint: "See MUD's version number and license information.", goDo: cmdLicense},
	"note":      {level: LEVEL_ANY, hint: "read notes", goDo: cmdNotes, args: []string{"note type", "list, next"}, modargs: []string{"note type", "create, setting"}},
	"crazytalk": {level: LEVEL_ANY, hint: "global chat with ascii-art text", goDo: cmdCrazyTalk, args: []string{"font", "message"}, hide: true, noAutoHelp: true},
	"charlist":  {level: LEVEL_ANY, hint: "see your list of characters", goDo: cmdCharList, noAutoHelp: true},
	"bug":       {level: LEVEL_ANY, hint: "Report a bug or typo in the game", goDo: cmdBug, args: []string{"report message"}, noAutoHelp: true},
	"logout":    {level: LEVEL_ANY, noShort: true, hint: "quit and go back to character selection menu", goDo: cmdLogout, noAutoHelp: true},
	"stats":     {level: LEVEL_ANY, hint: "show some mud stats", goDo: cmdStat, noAutoHelp: true},
	"recent":    {level: LEVEL_ANY, hint: "show recent logins", goDo: cmdLogins, noAutoHelp: true},

	//Builder/mod/imm
	"olc":      {level: LEVEL_BUILDER, noShort: true, hint: "world editor", goDo: cmdOLC, args: []string{"room,area,reset,object,mobile"}, noAutoHelp: true},
	"coninfo":  {level: LEVEL_MODERATOR, hint: "shows list of net connections", goDo: cmdConInfo, noAutoHelp: true},
	"pset":     {level: LEVEL_MODERATOR, hint: "set player parameters", goDo: cmdPset, args: []string{"target", "level", "level-number"}, noAutoHelp: true},
	"disable":  {level: LEVEL_MODERATOR, hint: "disable/enable a command or channel", goDo: cmdDisable, args: []string{"command/channel", "name of command or channel"}, noAutoHelp: true},
	"blocked":  {level: LEVEL_MODERATOR, hint: "Shows blocked connections", args: []string{"add or delete", "hostname or ip"}, goDo: cmdBlocked},
	"boom":     {level: LEVEL_MODERATOR, noShort: true, hint: "Boom a message at everyone", args: []string{"message"}, goDo: cmdBoom, noAutoHelp: true},
	"ban":      {level: LEVEL_MODERATOR, noShort: true, hint: "Ban a character", args: []string{"target", "reason"}, goDo: cmdBan, noAutoHelp: true},
	"unban":    {level: LEVEL_MODERATOR, noShort: true, hint: "Unban a character", args: []string{"target"}, goDo: cmdUnban, noAutoHelp: true},
	"settings": {level: LEVEL_MODERATOR, noShort: true, hint: "Change server settings", args: []string{"option"}, goDo: cmdServSet},
	"force":    {level: LEVEL_MODERATOR, noShort: true, hint: "Force a player to type something.", args: []string{"target/all", "command"}, goDo: cmdForce, noAutoHelp: true},
	"frecall":  {level: LEVEL_MODERATOR, hint: "Force a player to recall, for testing", args: []string{"target"}, goDo: cmdTransport},
	"panic":    {level: LEVEL_IMPLEMENTER, noShort: true, hint: "Test panic, recover, log and stackdump.", goDo: cmdPanic, noAutoHelp: true},
	"shutdown": {level: LEVEL_IMPLEMENTER, noShort: true, hint: "Shutdown", goDo: cmdShutdown, noAutoHelp: true},
}

func cmdOLC(player *characterData, input string) {
	interpOLC(player, input)
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

	player.send("%v averages:", durafmt.ParseShort(time.Duration(time.Duration(historyLen)/PULSE_PER_SECOND*time.Second)).Format(shortUnits))
	mp := (float64(ppulse) / float64(PULSE_LENGTH_uS)) * 100.0
	mpc := percentColor(mp)
	wp := (float64(fpulse) / float64(PULSE_LENGTH_uS)) * 100.0
	player.send("Mud load: Pulse: %v%3.4f%%{x / cmd-window: %3.2f%%", mpc, mp, wp)
	player.send("Pulse: %v",
		durafmt.ParseShort(time.Duration(ppulse*uint64(time.Microsecond))).Format(shortUnits))

	player.send("cmd-window: %v",
		durafmt.ParseShort(time.Duration(fpulse*uint64(time.Microsecond))).Format(shortUnits))

	fpulse = uint64(fullPulseHistory[historyLen-1])
	ppulse = uint64(partialPulseHistory[historyLen-1])
	player.send(NEWLINE + "Current:")
	mp = (float64(ppulse) / float64(PULSE_LENGTH_uS)) * 100.0
	mpc = percentColor(mp)
	wp = (float64(fpulse) / float64(PULSE_LENGTH_uS)) * 100.0
	player.send("Mud load: Pulse: %v%3.4f%%{x / cmd-window: %3.2f%%", mpc, mp, wp)
	player.send("Pulse: %v (%v peak)",
		durafmt.ParseShort(time.Duration(ppulse*uint64(time.Microsecond))).Format(shortUnits),
		durafmt.ParseShort(time.Duration(peakPartialPulse)*time.Microsecond).Format(shortUnits))

	player.send("cmd-window: %v (%v peak)",
		durafmt.ParseShort(time.Duration(fpulse*uint64(time.Microsecond))).Format(shortUnits),
		durafmt.ParseShort(time.Duration(peakFullPulse)*time.Microsecond).Format(shortUnits))
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
		player.Config.clearFlag(CONFIG_DEAF)
		player.send("You had the deaf option enabled, turning off.")
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

func cmdLogins(player *characterData, input string) {
	player.send("Recent logins:")

	if numRecentLogins == 0 {
		player.send("None.")
		return
	}
	for _, item := range recentLogins {
		player.send("%v: %v", item.Name, durafmt.Parse(time.Since(item.Time).Truncate(time.Second)).LimitFirstN(2).Format(shortUnits))
	}
}

func cmdWho(player *characterData, input string) {
	if player.Config.hasFlag(CONFIG_HIDDEN) {
		player.Config.clearFlag(CONFIG_HIDDEN)
		player.send("%v was enabled, turning off.", configNames[CONFIG_HIDDEN].name)
	}
	var buf string = "Players online:" + NEWLINE
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
	buf = buf + fmt.Sprintf("%31v - %v %v %v %v"+NEWLINE, "Player name", "level", "time-online", "(idle time)", "(no link)")
	for _, target := range tmpCharList {
		hidden := ""
		if target.Config.hasFlag(CONFIG_HIDDEN) {
			if player.Level >= LEVEL_BUILDER {
				hidden = " (hidden)"
			} else {
				continue
			}
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
		buf = buf + fmt.Sprintf("%31v - %v %v%v%v%v"+NEWLINE, target.Name, levelToName[target.Level].Name, onlineTime, idleTime, unlink, hidden)
	}
	uptime := durafmt.Parse(time.Since(bootTime).Truncate(time.Second)).LimitFirstN(2).Format(shortUnits)
	uptime = strings.ReplaceAll(uptime, " ", "")
	buf = buf + fmt.Sprintf(NEWLINE+"Players: %v (most: %v), Logins: %v (ever: %v), Uptime: %v", numPlayers, mudStats.MostEver, mudStats.loginCount, mudStats.LoginEver, uptime)

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
