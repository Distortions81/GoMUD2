package main

import (
	"strconv"
	"strings"
)

// DO NOT CHANGE ORDER
const (
	CONFIG_HIDDEN = 1 << iota
	CONFIG_NOTELL
	CONFIG_NOCHANNEL
	CONFIG_DEAF
	CONFIG_NOWRAP
	CONFIG_NOCOLOR
	CONFIG_TEXT_EMOJI
	CONFIG_OLC
	CONFIG_OLCHERE
	CONFIG_OLCHYBRID
	CONFIG_TERMWIDTH

	//Keep at end, do not use or delete
	CONFIG_MAX
)

type configInfo struct {
	name, desc string
	level      int

	integer      bool
	defaultValue int

	disableWhenEnabled,
	enableWhenEnabled,
	disableWhenDisabled,
	enableWhenDisabled int
}

var configNames map[int]configInfo = map[int]configInfo{
	CONFIG_HIDDEN:     {name: "Hide", desc: "Don't show up in who, join or leave."},
	CONFIG_NOTELL:     {name: "MuteTells", desc: "Reject tells"},
	CONFIG_NOCHANNEL:  {name: "MuteChannels", desc: "Mute all channels"},
	CONFIG_DEAF:       {name: "Deaf", desc: "Mute say/emote/yell"},
	CONFIG_NOWRAP:     {name: "NoWrap", desc: "Do not word-wrap"},
	CONFIG_NOCOLOR:    {name: "NoColor", desc: "Disable ANSI color"},
	CONFIG_TEXT_EMOJI: {name: "TextEmoji", desc: "Attempt to replace emoji with emoji names."},
	CONFIG_OLCHERE:    {name: "OLCHere", desc: "Always edit current area/room by default", level: LEVEL_BUILDER},
	CONFIG_OLC:        {name: "OLCMode", desc: "Require 'OLC' before OLC commands.", level: LEVEL_BUILDER, disableWhenEnabled: CONFIG_OLCHYBRID},
	CONFIG_OLCHYBRID:  {name: "OLCHybrid", desc: "Allow OLC and normal commands at the same time.", level: LEVEL_BUILDER, disableWhenEnabled: CONFIG_OLC},
	CONFIG_TERMWIDTH:  {name: "TermWidth", desc: "Manually specify your terminal width in columns.", integer: true, defaultValue: 80},
}

func cmdConfig(player *characterData, input string) {
	if input == "" {
		for x := 0; x < CONFIG_MAX; x++ {
			item := configNames[1<<x]
			if item.level > player.Level {
				continue
			}
			if item.name == "" {
				continue
			}

			if item.integer {
				if !player.Config.hasFlag(1 << x) {
					player.send("%15v: (%v) %v", cEllip(item.name, 15), boolToText(false), item.desc)
				} else {
					value := 0
					if player.ConfigVals[1<<x] != nil {
						value = player.ConfigVals[1<<x].Value
					}
					player.send("%15v: (%v) %v", cEllip(item.name, 15), value, item.desc)
				}
				continue
			}

			status := boolToText(player.Config.hasFlag(1 << x))
			player.send("%15v: (%v) %v", cEllip(item.name, 15), status, item.desc)
		}
		player.send("config <option> to toggle")
		return
	}

	parts := strings.Split(input, " ")
	numParts := len(parts)
	found := false
	for y := 0; y < numParts; y++ {
		for x := 0; x < CONFIG_MAX; x++ {
			item := configNames[1<<x]
			if item.level > player.Level {
				continue
			}
			if item.name == "" {
				continue
			}
			if strings.EqualFold(item.name, parts[y]) {
				found = true

				if item.integer && y < numParts-1 {
					i, err := strconv.ParseInt(parts[y+1], 10, 64)
					if err == nil {
						if i == 0 {
							delete(player.ConfigVals, 1<<x)
						} else {
							player.ConfigVals[1<<x] = &ConfigValue{Name: item.name, Value: int(i)}
						}

						player.send("%v is now %v.", item.name, i)
						player.Config.addFlag(1 << x)
						continue
					}
				}
				if player.Config.hasFlag(1 << x) {
					player.send("%v is now OFF.", item.name)
					player.Config.clearFlag(1 << x)

					if item.integer {
						if player.ConfigVals[1<<x] != nil {
							delete(player.ConfigVals, 1<<x)
						}
					}

					if item.disableWhenDisabled > 0 {
						player.Config.clearFlag(Bitmask(item.disableWhenEnabled))
					}
					if item.enableWhenDisabled > 0 {
						player.Config.addFlag(Bitmask(item.enableWhenEnabled))
					}
				} else {
					player.send("%v is now ON", item.name)
					player.Config.addFlag(1 << x)

					if item.integer {
						if player.ConfigVals[1<<x] == nil {
							player.ConfigVals[1<<x] = &ConfigValue{Name: item.name, Value: int(item.defaultValue)}
						}
					}

					if item.disableWhenEnabled > 0 {
						player.Config.clearFlag(Bitmask(item.disableWhenEnabled))
					}
					if item.enableWhenEnabled > 0 {
						player.Config.addFlag(Bitmask(item.enableWhenEnabled))
					}

				}
				player.dirty = true
				break
			}
		}
		if !found {
			player.send("I don't see a config option with that name.")
		}
	}
}
