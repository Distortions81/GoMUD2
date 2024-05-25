package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

var attemptMap map[string]*attemptData = make(map[string]*attemptData)

type attemptData struct {
	Host     string
	Attempts int
	Blocked  bool
	HTTP     bool

	Created  time.Time
	Modified time.Time
}

func cmdBlocked(player *characterData, input string) {
	args := strings.Split(input, " ")
	numArgs := len(args)
	var target string

	if strings.EqualFold(input, "clear") {
		if len(attemptMap) > 0 {
			attemptMap = make(map[string]*attemptData)
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
				if attemptMap[target] != nil {
					player.send("The host '%v' has been deleted from the list.", target)
					delete(attemptMap, target)
				} else {
					player.send("The host '%v' was not found in the list.", target)
				}
			} else if strings.EqualFold(args[0], "add") {

				if attemptMap[target] == nil {
					attemptMap[target] = &attemptData{Host: target, Blocked: true, Created: time.Now()}
					player.send("Host '%v' added to the list.", target)
				} else {
					if attemptMap[target].Blocked {
						player.send("Host '%v' was already blocked.", target)
					} else {
						attemptMap[target].Blocked = true
						player.send("The host '%v' was already in the list. Blocking it.", target)
					}
				}
				attemptMap[target].Modified = time.Now()
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

	var atd []*attemptData
	for i, item := range attemptMap {
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
