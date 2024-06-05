package main

import (
	"fmt"
	"strings"
)

const (
	vnumSpace = 1
	UNDO_MAX  = 50
)

const (
	OLC_NONE = iota
	OLC_ROOM
	OLC_AREA
	OLC_RESET
	OLC_OBJECT
	OLC_MOBILE

	OLC_MAX
)

type olcModeType struct {
	name string
	list []*commandData
}

var modeToText [OLC_MAX]string

var olcModes [OLC_MAX]olcModeType = [OLC_MAX]olcModeType{
	OLC_NONE:   {name: "NONE"},
	OLC_ROOM:   {name: "ROOM", list: roomCmds},
	OLC_AREA:   {name: "AREA"},
	OLC_RESET:  {name: "RESET"},
	OLC_OBJECT: {name: "OBJECT"},
	OLC_MOBILE: {name: "MOBILE"},
}

// Available in all modes
var gOLCcmds []*commandData = []*commandData{
	{name: "cmd", goDo: olcExternalCmd, hint: "run a normal non-OLC command (see config OLCMode)", args: []string{"command"}},
	{name: "exit", goDo: olcExit, hint: "exit OLC"},
	{name: "asave", goDo: olcAsaveAll, hint: "save all areas"},
	{name: "help", subType: olcHelp, hint: "list available commands"},
}

func olcExit(player *characterData, input string) {
	player.OLCEditor.OLCMode = OLC_NONE
	player.send("Exited OLC mode.")
}

func olcExternalCmd(player *characterData, input string) {
	parseCommand(player, input)
}

func olcHelp(player *characterData, olcCmdList []*commandData) {
	for _, item := range olcCmdList {
		var parts string
		for i, arg := range item.args {
			if i > 0 {
				parts += ", "
			}
			parts += fmt.Sprintf("<%v>", arg)
		}
		if parts != "" {
			parts += " "
		}
		player.sendWW("%10v %v-- %v", cEllip(item.name, 10), parts, item.hint)
	}
}

func olcModeCommand(mode olcModeType, player *characterData, input string) bool {
	args := strings.SplitN(input, " ", 2)

	for _, item := range mode.list {
		if strings.EqualFold(item.name, args[0]) {
			if item.goDo == nil {
				item.subType(player, mode.list)
				return false
			}
			if len(args) != 2 {
				item.goDo(player, "")
			} else {
				item.goDo(player, args[1])
			}
			return true
		}
	}
	if len(args) != 2 {
		if !findCommandMatch(mode.list, player, args[0], "") {
			if !player.Config.hasFlag(CONFIG_OLCHYBRID) {
				olcHelp(player, mode.list)
				return true
			}
			return false
		}
	} else {
		if !findCommandMatch(mode.list, player, args[0], args[1]) {
			if !player.Config.hasFlag(CONFIG_OLCHYBRID) {
				olcHelp(player, mode.list)
				return true
			}
			return false
		}
	}
	return true
}

func interpOLC(player *characterData, input string) bool {

	if input == "" {
		for i, item := range olcModes {
			if i == player.OLCEditor.OLCMode {
				olcHelp(player, item.list)
				return true
			}
		}
	}
	if player.OLCEditor.OLCMode != OLC_NONE {

		if strings.EqualFold("exit", input) {
			player.OLCEditor.OLCMode = OLC_NONE
			player.send("Exited OLC editor.")
			return true
		}
		for i, item := range olcModes {
			if i == player.OLCEditor.OLCMode {
				if olcModeCommand(item, player, input) {
					return true
				}
			}
		}
	}
	for i, item := range olcModes {
		if strings.EqualFold(item.name, input) {
			player.OLCEditor.OLCMode = i
			player.send("Entering %v edit mode.", item.name)
			return true
		}
	}
	if !player.Config.hasFlag(CONFIG_OLCHYBRID) {
		player.send("That isn't a valid OLC command.")

		player.send("OLC modes:")
		for _, item := range olcModes {
			player.send("%-10v", cEllip(item.name, 10))
		}
		return true
	}

	return false
}

func olcAsaveAll(player *characterData, input string) {
	if player.Level < LEVEL_BUILDER {
		return
	}
	saveAllAreas(true)
	player.send("all areas saved.")
}
