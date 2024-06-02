package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goMUD2/figletlib"
	"strconv"
	"strings"
	"time"

	"github.com/hako/durafmt"
)

type serverSettingsData struct {
	Version          int
	NewLock, ModOnly bool
}

var servSet serverSettingsData

func cmdServSet(player *characterData, input string) {

	if input == "" {
		player.send("NewLock: %v, ModOnly: %v",
			boolToText(servSet.ModOnly), boolToText(servSet.NewLock))
		player.send("settings: <option name> toggles specified setting on/off")
		return
	}

	if strings.EqualFold(input, "modonly") {
		if servSet.ModOnly {
			servSet.ModOnly = false
		} else {
			servSet.ModOnly = true
		}
		servSet.saveSettings()
		cmdServSet(player, "")
		player.send("ModeratorOnly is now %v.", boolToText(servSet.ModOnly))

	} else if strings.EqualFold(input, "newlock") {
		if servSet.NewLock {
			servSet.NewLock = false
		} else {
			servSet.NewLock = true
		}
		servSet.saveSettings()
		cmdServSet(player, "")
		player.send("NewLock is now %v.", boolToText(servSet.NewLock))

	} else {
		player.send("That isn't a valid option.")
	}
}
func cmdForce(player *characterData, input string) {
	args := strings.SplitN(input, " ", 2)

	if input == "" {
		player.send("force <player name/all> <command>")
	}
	if len(args) < 2 {
		player.send("But what command?")
		return
	}
	if strings.EqualFold(args[1], "force") {
		player.send("You can't use force to run force.")
		return
	}

	if strings.EqualFold(args[0], "all") {
		for _, target := range charList {
			if target == player {
				//Don't force yourself
				continue
			}
			goForce(target, args[1])
			target.send("%v forced you to: %v", player.Name, args[1])
		}
		player.send("Forced everyone to: %v", args[1])
		critLog("%v forced everyone to: %v", player.Name, args[1])
		return
	}

	if target := checkPlaying(args[0]); target != nil {
		if target == player {
			player.send("You can't force youself")
			return
		}
		goForce(target, args[1])
		target.send("%v forced you to: %v", player.Name, args[1])
		player.send("You forced %v to: %v", target.Name, args[1])
		critLog("%v forced %v to: %v", player.Name, target.Name, args[1])

	} else {
		player.send("They don't seem to be online.")
	}
}

func goForce(player *characterData, input string) {
	cmdStr, args, _ := strings.Cut(input, " ")
	cmdStr = strings.ToLower(cmdStr)

	var command *commandData
	for _, cmd := range cmdList {
		if strings.EqualFold(cmd.name, cmdStr) {
			command = cmd
			break
		}
	}
	if command != nil {
		if command.disabled {
			player.send("That command is disabled.")
			return
		}
		if command.checkCommandLevel(player) {
			if command.forceArg != "" {
				command.goDo(player, command.forceArg)
			} else {
				command.goDo(player, args)
			}
		}
	} else {
		if cmdChat(player, input) {
			findCommandMatch(player, cmdStr, args)
		}
	}
}

func loadSettings() serverSettingsData {
	set := serverSettingsData{}

	data, err := readFile(DATA_DIR + SETTINGS_FILE)
	if err != nil {
		return set
	}

	err = json.Unmarshal(data, &set)
	if err != nil {
		critLog("loadPlayer: Unable to unmarshal the data.")
		return set
	}
	return set
}

func (set *serverSettingsData) saveSettings() {
	fileName := DATA_DIR + SETTINGS_FILE
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	set.Version = SETTINGS_VERSION
	err := enc.Encode(&set)
	if err != nil {
		critLog("saveSettings: enc.Encode: %v", err.Error())
		return
	}

	err = saveFile(fileName, outbuf.Bytes())
	if err != nil {
		critLog("saveSettings: saveFile failed %v", err.Error())
		return
	}
}

func cmdUnban(player *characterData, input string) {
	doBan(player, input, true, false)
}
func cmdBan(player *characterData, input string) {
	doBan(player, input, false, false)
}

func doBan(player *characterData, input string, unban, account bool) {
	args := strings.SplitN(input, " ", 2)
	argCount := len(args)

	if input == "" {
		player.send("Ban a character: ban <player> <reason>")
		return
	}

	reason := "No reason given."
	if argCount == 2 {
		r := strings.TrimSpace(args[1])
		if len(r) > 1 {
			reason = r
		}
	}

	var target *characterData
	if target = checkPlaying(args[0]); target != nil {
		target.send("You have been banned. Reason: %v", reason)
		target.desc.close()
	}

	if target == nil {
		tDesc := descData{}
		if target = tDesc.pLoad(args[0]); target == nil {
			player.send("Unable to find a player by that name.")
			return
		}
	}

	foundBan := false
	if unban {
		for i, item := range target.Banned {
			if !item.Revoked {
				target.Banned[i].Revoked = true
				foundBan = true
			}
		}
		if foundBan {
			player.send("%v has been unbanned.", target.Name)
			critLog("%v has been unbanned by %v.", target.Name, player.Name)

			target.saveCharacter()
			target.valid = false
		} else {
			player.send("%v wasn't banned.", target.Name)
		}
		return
	}

	target.Banned = append(target.Banned, banData{Reason: reason, Date: time.Now().UTC(), BanBy: player.Name})
	target.saveCharacter()
	player.send("%v has been banned: %v", target.Name, reason)
	critLog("%v has been banned: %v: %v", target.Name, player.Name, reason)
	target.quit(true)
	//player.sendToPlaying(" --> %v has been banned. <--", target.Name)
}

func cmdBoom(player *characterData, input string) {
	buf := fmt.Sprintf("%v booms: %v", player.Name, input)
	boom, err := figletlib.TXTToAscii(buf, "standard", "left", 0)
	if err != nil {
		player.send("Sorry, unable to load the font.")
		return
	}
	for _, target := range charList {
		target.send(boom)
	}
}

func cmdConInfo(player *characterData, input string) {
	player.send("Descriptors:")
	for _, item := range descList {
		player.send("\r\nID: %-32v IP: %v", item.id, item.ip)
		player.send("State: %-29v DNS: %v", stateName[item.state], item.dns)
		player.send("Idle: %-30v Connected: %v", durafmt.ParseShort(time.Since(item.idleTime)), durafmt.ParseShort(time.Since(item.connectTime)))

		charmap := item.telnet.charMap.String()
		if item.telnet.Options != nil && item.telnet.Options.UTF {
			charmap = "UTF"
		}

		player.send("Clinet: %-28v Charmap: %v", item.telnet.termType, charmap)
		if item.character != nil {
			player.send("Char: %-30v Account: %v", item.character.Name, item.account.Login)
		}
	}
}

func cmdPset(player *characterData, input string) {
	var target *characterData

	if input == "" {
		cmdHelp(player, "pset")
	}

	args := strings.Split(input, " ")
	numArgs := len(args)

	name := args[0]
	if target = checkPlaying(name); target == nil {
		player.send("They aren't online at the moment.")
		return
	}

	if numArgs < 2 {
		cmdHelp(player, "pset")
		return
	}

	command := strings.ToLower(args[1])
	if command == "level" {
		if numArgs < 3 {
			cmdHelp(player, "pset")
			return
		}
		level, err := strconv.Atoi(args[2])
		if err != nil {
			player.send("That isn't a number.")
			return
		} else {
			if level > player.Level {
				player.send("You can't set a player's level to a level higher than your own.")
				return
			}
			target.Level = level
			player.send("%v's level is now %v.", target.Name, target.Level)
			target.dirty = true
			return
		}
	}

}

func cmdTransport(player *characterData, input string) {

	if input == "" {
		player.send("Send who where?")
		return
	}
	if target := checkPlayingPMatch(input); target != nil {
		target.leaveRoom()
		target.sendToRoom("%v suddenly vanishes!", target.Name)
		target.send("You have been forced to recall.")
		target.goTo(LocData{AreaUUID: sysAreaUUID, RoomUUID: sysRoomUUID})
		player.send("Forced %v to recall.", target.Name)
		return
	}
	player.send("I don't see anyone by that name.")
}
