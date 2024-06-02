package main

import (
	"fmt"
	"goMUD2/figletlib"
	"strconv"
	"strings"
	"time"

	"github.com/hako/durafmt"
)

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
	cmd := strings.SplitN(input, " ", 2)
	cmdLen := len(cmd)

	if input == "" {
		player.send("Send who where?")
		return
	}
	if cmdLen == 1 {
		player.send("Send them where?")
	}
	if cmdLen == 2 {
		if target := checkPlayingPMatch(cmd[0]); target != nil {
			target.leaveRoom()
			target.send("You have been forced to recall.")
			target.goTo(LocData{AreaUUID: sysAreaUUID, RoomUUID: sysRoomUUID})
			player.send("Sent %v to %v.")
			return
		}
		player.send("I don't see anyone by that name.")
	}
}
