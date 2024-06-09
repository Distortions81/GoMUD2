package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

var disabled disabledData

type disabledData struct {
	Version int
	Items   []disabledItemData
}

type disabledItemData struct {
	Kind string
	Name string
}

func updateDisables() {
	disabled.Items = []disabledItemData{}
	for _, item := range cmdList {
		if item.disabled {
			disabled.Items = append(disabled.Items, disabledItemData{Kind: "command", Name: item.name})
		}
	}
	for _, item := range channels {
		if item.disabled {
			disabled.Items = append(disabled.Items, disabledItemData{Kind: "channel", Name: item.name})
		}
	}
}

func restoreDisables() {
	for _, cmd := range cmdList {
		cmd.disabled = false
	}
	for _, ch := range channels {
		ch.disabled = false
	}
	for _, item := range disabled.Items {
		if item.Kind == "command" {
			for _, cmd := range cmdList {
				if cmd.name == item.Name {
					cmd.disabled = true
				}
			}
		} else if item.Kind == "channel" {
			for _, ch := range channels {
				if ch.name == item.Name {
					ch.disabled = true
				}
			}
		}
	}
}

func writeDisables() {
	updateDisables()

	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	disabled.Version = DISABLES_VERSION

	err := enc.Encode(&disabled)
	if err != nil {
		critLog("writeDisables: enc.Encode: %v", err.Error())
		return
	}

	err = saveFile(DATA_DIR+DISABLES_FILE, outbuf.Bytes())
	if err != nil {
		critLog("writeDisables: saveFile failed %v", err.Error())
		return
	}
}

func readDisables() {
	data, err := readFile(DATA_DIR + DISABLES_FILE)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &disabled)
	if err != nil {
		critLog("readDisables: Unable to unmarshal the data.")
		return
	}
	restoreDisables()
}

const disCol = 4

func cmdDisable(player *characterData, input string) {
	parts := strings.SplitN(input, " ", 2)
	numParts := len(parts)

	if input == "" {
		player.send("Commands:")
		buf := ""
		count := 1
		for _, cmd := range cmdList {
			if cmd.hide {
				continue
			}

			buf = buf + fmt.Sprintf("%10v ", cEllip(cmd.name, 10))
			if cmd.disabled {
				buf = buf + "(X) "
			} else {
				buf = buf + "( ) "
			}
			if count%disCol == 0 {
				buf = buf + NEWLINE
				count = 0
			}
			count++
		}
		player.send(buf)

		player.send(NEWLINE + "Channels:")
		buf = ""
		count = 1
		for _, cmd := range channels {
			buf = buf + fmt.Sprintf("%10v ", cEllip(cmd.cmd, 10))
			if cmd.disabled {
				buf = buf + "(X)"
			} else {
				buf = buf + "( )"
			}
			if count%disCol == 0 {
				buf = buf + NEWLINE
				count = 0
			}
			count++
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
					if !cmd.disabled {
						cmd.disabled = true
						player.send("The %v command is now {rdisabled{x.", cmd.name)
					} else {
						cmd.disabled = false
						player.send("The %v command is now {genabled{x.", cmd.name)
					}
					writeDisables()
					return
				}
			}
			player.send("I don't see a command called that.")
			return
		} else if strings.EqualFold(parts[0], "channel") {
			for _, ch := range channels {
				if strings.EqualFold(ch.cmd, parts[1]) {
					if !ch.disabled {
						ch.disabled = true
						player.send("The %v channel is now {rdisabled{x.", ch.name)
					} else {
						ch.disabled = false
						player.send("The %v channel is now {genabled{x.", ch.name)
					}
					writeDisables()
					return
				}
			}
			player.send("I don't see a channel called that.")
			return
		}
	}
	player.send("Disable a command or a channel?")
}
