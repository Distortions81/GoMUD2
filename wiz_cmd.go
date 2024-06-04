package main

import (
	"fmt"
	"goMUD2/figletlib"
	"strconv"
	"strings"
	"time"

	"github.com/hako/durafmt"
)

func cmdPanic(player *characterData, input string) {
	panic("test")
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

/* To do: remove dupe code */
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
			if !findCommandMatch(cmdList, player, cmdStr, args) {
				player.listCommands("")
			}
		}
	}
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
		player.send("State: %-29v DNS: %v", cEllip(stateName[item.state], 29), item.dns)
		player.send("Idle: %-30v Connected: %v",
			cEllip(durafmt.ParseShort(time.Since(item.idleTime)).String(), 30),
			durafmt.ParseShort(time.Since(item.connectTime)))

		charmap := item.telnet.charMap.String()
		if item.telnet.Options != nil && item.telnet.Options.UTF {
			charmap = "UTF"
		}

		player.send("Clinet: %-28v Charmap: %v", cEllip(item.telnet.termType, 28), charmap)
		if item.character != nil {
			player.send("Char: %-30v Account: %v", cEllip(item.character.Name, 30), item.account.Login)
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
