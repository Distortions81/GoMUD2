package main

import (
	"sort"
	"strconv"
	"strings"
)

const disCol = 4

func cmdConInfo(player *characterData, input string) {
	player.send("Characters:")
	for _, item := range charList {
		if item.desc != nil {
			player.send("Name: %30v -- ID: %v", item.Name, item.desc.id)
		} else {
			player.send("Name: %30v -- (no link)", item.Name)
		}
	}
	player.send("\r\nDescriptors:")
	for _, item := range descList {
		player.send("id: %04v addr: %16v state: %v\r\ndns: %v ", item.id, item.ip, item.state, item.dns)
	}
}

type attemptData struct {
	Attempts int
	Host     string
}

func cmdBlocked(player *characterData, input string) {
	args := strings.Split(input, " ")
	numArgs := len(args)
	var target string

	if strings.EqualFold(input, "clear") {
		if len(attemptMap) > 0 {
			attemptMap = make(map[string]int)
			player.send("The block list has been cleared.")
		} else {
			player.send("The list is already empty.")
		}
		return
	}
	if numArgs > 1 {
		target = args[1]
		if target != "" {
			if strings.EqualFold(args[0], "delete") {
				if attemptMap[target] != 0 {
					player.send("%v has been deleted from the list.", target)
					delete(attemptMap, target)
				}
			} else if strings.EqualFold(args[0], "add") {
				if attemptMap[target] == 0 {
					player.send("%v added to the list.", target)
					attemptMap[target] = -1
				}
			} else if args[0] == "" {
				player.send("Delete, or add item?")
			}
		} else {
			player.send("But what host?")
		}
		return
	} else if input != "" {
		player.send("But what host?")
		return
	}

	var atd []attemptData
	for i, item := range attemptMap {
		atd = append(atd, attemptData{Attempts: item, Host: i})
	}

	sort.Slice(atd, func(i, j int) bool {
		return atd[i].Attempts < atd[j].Attempts
	})

	player.send("Blocked connections: host: attempts or blocked")

	count := 0
	for _, item := range atd {
		if item.Attempts == 0 {
			continue
		}
		count++
		if item.Attempts == -1 {
			player.send("%70v : %v", item.Host, "Blocked")
		} else {
			player.send("%70v : %v", item.Host, item.Attempts)
		}
	}
	if count == 0 {
		player.send("There are no blocked connections.")
	} else {
		player.send("Type 'blocked clear' to clear the list... or <add or delete> <host>")
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
