package main

import (
	"fmt"
	"strconv"
	"strings"
)

const disCol = 4

func cmdDisable(player *characterData, input string) {
	parts := strings.SplitN(input, " ", 2)
	numParts := len(parts)

	if input == "" {
		player.send("Commands:")
		var buf string
		var count int
		var c int
		for _, cmd := range cmdList {
			if cmd.hide {
				continue
			}

			if c%disCol == 0 {
				buf = buf + "\r\n"
				count = 0
			}
			if count > 0 {
				buf = buf + ", "
			}
			if cmd.Disabled {
				buf = buf + "(X) "
			} else {
				buf = buf + "( ) "
			}
			buf = buf + fmt.Sprintf("%10v", cmd.name)
			count++
			c++
		}
		player.send(buf)

		player.send("\r\nChannels:")
		buf = ""
		count = 0
		c = 0
		for _, cmd := range channels {
			if c%disCol == 0 {
				buf = buf + "\r\n"
				count = 0
			}
			if count > 0 {
				buf = buf + ", "
			}
			if cmd.disabled {
				buf = buf + "(X) "
			} else {
				buf = buf + "( ) "
			}
			buf = buf + fmt.Sprintf("%10v", cmd.cmd)
			count++
			c++
		}
		player.send(buf)
		return
	}
	if numParts == 2 {
		if strings.EqualFold(parts[0], "command") {
			if strings.EqualFold(parts[1], "disable") {
				player.send("You can't disable the disable command.")
				return
			}
			for _, cmd := range cmdList {
				if strings.EqualFold(cmd.name, parts[1]) {
					if !cmd.Disabled {
						cmd.Disabled = true
						player.send("The %v command is now disabled.", cmd.name)
					} else {
						cmd.Disabled = false
						player.send("The %v command is now enabled.", cmd.name)
					}
					return
				}
			}
			player.send("I don't see a command called that.")
			return
		} else if strings.EqualFold(parts[0], "channel") {
			for _, ch := range channels {
				if strings.EqualFold(ch.cmd, parts[1]) {
					ch.disabled = true
					if !ch.disabled {
						ch.disabled = true
						player.send("The %v channel is now disabled.", ch.name)
					} else {
						ch.disabled = false
						player.send("The %v channel is now enabled.", ch.name)
					}
					return
				}
			}
			player.send("I don't see a channel called that.")
			return
		}
	}
	player.send("Disable a command or a channel?")
}

func cmdConInfo(player *characterData, input string) {
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
