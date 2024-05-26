package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	BLOCKED_SAVE_INTERVAL = time.Minute
	BLOCKED_THRESH        = 6
	BLOCKED_FORGIVE       = 15
	BLOCKED_COOLDOWN      = time.Hour
)

var blockedMap map[string]*blockedData = make(map[string]*blockedData)
var blockedDirty bool
var lastBlockedSave time.Time

type blockedData struct {
	Host     string
	Attempts int  `json:",omitempty"`
	Blocked  bool `json:",omitempty"`
	HTTP     bool `json:",omitempty"`

	Created  time.Time
	Modified time.Time
}

func expireBlocks() {
	for i, item := range blockedMap {
		if item.HTTP {
			continue
		}
		if !item.Blocked {
			continue
		}
		if item.Attempts > BLOCKED_FORGIVE {
			continue
		}
		if time.Since(item.Modified) < BLOCKED_COOLDOWN {
			continue
		}
		errLog("Removing block on '%v' because of cooldown.", i)
		delete(blockedMap, i)
		break
	}
}

func writeBlocked(force bool) {

	if !force && !blockedDirty {
		return
	}
	if time.Since(lastBlockedSave) < SAVE_INTERVAL {
		return
	}
	lastBlockedSave = time.Now()

	var atd []blockedData
	for _, item := range blockedMap {
		atd = append(atd, *item)
	}

	sort.Slice(atd, func(i, j int) bool {
		return atd[i].Attempts < atd[j].Attempts || atd[i].Host < atd[j].Host
	})

	go func(atd []blockedData) {
		fileName := DATA_DIR + BLOCKED_FILE
		outbuf := new(bytes.Buffer)
		enc := json.NewEncoder(outbuf)
		enc.SetIndent("", "\t")

		err := enc.Encode(&atd)
		if err != nil {
			critLog("writeBlocked: enc.Encode: %v", err.Error())
			return
		}

		err = saveFile(fileName, outbuf.Bytes())
		if err != nil {
			critLog("writeBlocked: saveFile failed %v", err.Error())
			return
		}
	}(atd)
	blockedDirty = false
}

func readBlocked() error {
	lastBlockedSave = time.Now()

	fileName := DATA_DIR + BLOCKED_FILE
	data, err := readFile(fileName)
	if err != nil {
		return err
	}

	var blocked []blockedData
	err = json.Unmarshal(data, &blocked)
	if err != nil {
		critLog("readBlocked: Unable to unmarshal the data.")
		return err
	}

	blockedMap = make(map[string]*blockedData)
	for _, item := range blocked {
		blockedMap[item.Host] = &blockedData{
			Host: item.Host, Attempts: item.Attempts,
			Blocked: item.Blocked, HTTP: item.HTTP,
			Created: item.Created, Modified: item.Modified}
	}
	return nil
}

func cmdBlocked(player *characterData, input string) {
	args := strings.Split(input, " ")
	numArgs := len(args)
	var target string

	if strings.EqualFold(input, "clear") {
		if len(blockedMap) > 0 {
			blockedMap = make(map[string]*blockedData)
			player.send("The block list has been cleared.")
			blockedDirty = true
		} else {
			player.send("The list is already empty.")
		}
		return
	}
	if numArgs > 1 {
		target = args[1]
		if target != "" {
			if strings.EqualFold(args[0], "delete") {
				if blockedMap[target] != nil {
					player.send("The host '%v' has been deleted from the list.", target)
					delete(blockedMap, target)
					blockedDirty = true
				} else {
					player.send("The host '%v' was not found in the list.", target)
				}
			} else if strings.EqualFold(args[0], "add") {

				if blockedMap[target] == nil {
					blockedMap[target] = &blockedData{Host: target, Blocked: true, Created: time.Now()}
					player.send("Host '%v' added to the list.", target)
					blockedDirty = true
				} else {
					if blockedMap[target].Blocked {
						player.send("Host '%v' was already blocked.", target)
					} else {
						blockedMap[target].Blocked = true
						player.send("The host '%v' was already in the list. Blocking it.", target)
						blockedDirty = true
					}
				}
				blockedMap[target].Modified = time.Now()
				blockedDirty = true
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

	var atd []*blockedData
	for i, item := range blockedMap {
		item.Host = i
		atd = append(atd, item)
	}

	sort.Slice(atd, func(i, j int) bool {
		return atd[i].Attempts < atd[j].Attempts || atd[i].Host < atd[j].Host
	})

	player.send("Blocked connections:")
	player.send("%50v : %-8v %v %v", "hostname", "Attempts", "(Blocked)", "(HTTP)")

	count := 0
	var buf string
	for _, item := range atd {
		if item.Attempts == 0 && !item.Blocked {
			continue
		}
		count++
		buf = buf + fmt.Sprintf("%50v : %8v", item.Host, item.Attempts)
		if item.Blocked {
			buf = buf + " (Blocked)"
		}
		if item.HTTP {
			buf = buf + " (HTTP)"
		}
		buf = buf + "\r\n"
	}
	player.send(buf)
	if count == 0 {
		player.send("There are no blocked connections.")
	}
	player.send("Type 'blocked clear' to clear the list... or <add or delete> <host>")

}