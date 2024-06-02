package main

import "strings"

// DO NOT CHANGE ORDER
const (
	CONFIG_HIDDEN = 1 << iota
	CONFIG_NOTELL
	CONFIG_NOCHANNEL
	CONFIG_DEAF
	CONFIG_NOWRAP
	CONFIG_OLC
	CONFIG_NOCOLOR
	CONFIG_TEXT_EMOJI

	//Keep at end, do not use or delete
	CONFIG_MAX
)

type configInfo struct {
	name, desc string
	level      int
}

var configNames map[int]configInfo = map[int]configInfo{
	CONFIG_HIDDEN:     {name: "Hidden", desc: "Don't show up in who, join or leave."},
	CONFIG_NOTELL:     {name: "NoTell", desc: "Reject tells"},
	CONFIG_NOCHANNEL:  {name: "NoChannel", desc: "Mute all channels"},
	CONFIG_DEAF:       {name: "Deaf", desc: "Mute say/emote/yell"},
	CONFIG_NOWRAP:     {name: "NoWrap", desc: "Do not word-wrap"},
	CONFIG_NOCOLOR:    {name: "NoColor", desc: "Disable ANSI color"},
	CONFIG_OLC:        {name: "OLCMode", desc: "Require 'OLC' before OLC commands.", level: LEVEL_BUILDER},
	CONFIG_TEXT_EMOJI: {name: "TextEmoji", desc: "Attempt to replace emoji with emoji names.", level: LEVEL_BUILDER},
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

			status := boolToText(player.Config.hasFlag(1 << x))
			player.send("%15v: (%v) %v", item.name, status, item.desc)
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

				if player.Config.hasFlag(1 << x) {
					player.send("%v is now OFF.", item.name)
					player.Config.clearFlag(1 << x)
				} else {
					player.send("%v is now ON", item.name)
					player.Config.addFlag(1 << x)
				}
				if player.desc != nil && player.desc.telnet.Options != nil {
					player.desc.telnet.Options.NoColor = player.Config.hasFlag(CONFIG_NOCOLOR)
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
