package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

var helpFiles []*helpTopicData

type helpTopicData struct {
	Topic             string
	Created, Modified time.Time

	Helps []helpData
	dirty bool
}

type helpData struct {
	Name              string
	Description       string
	Created, Modified time.Time

	Keywords []string
	Authors  []string

	Text  string
	topic *helpTopicData
}

func cmdHelp(player *characterData, input string) {
	//Redirect command list
	if player.desc != nil && strings.EqualFold("commands", input) {
		player.listCommands("")
		return
	}

	//Autohelp
	for _, item := range cmdList {
		if strings.EqualFold(input, item.name) {
			player.send("Help %v:", item.name)
			if !item.noShort && !item.noAutoHelp {
				item.goDo(player, "")
				return
			} else {
				player.listCommands(item.name)
				return
			}
		}
	}

	if player.desc != nil && strings.EqualFold("emoji", input) {
		player.send(emojiHelp)
		return
	}

	if player.desc != nil && strings.EqualFold("more-emoji", input) {
		player.send(moreEmojiHelp)
		return
	}

	count := 0
	buf := ""
	//Search help topics
	for _, topic := range helpFiles {
		if topic.Topic == input {
			for _, help := range topic.Helps {
				if count > 0 {
					buf = buf + ", "
				}
				buf = buf + help.Name
				count++
			}
			if count > 0 {
				player.send("Found these help topics:\r\n" + buf)
				return
			}
		}
	}
	//Search individual help entries
	for _, topic := range helpFiles {
		for _, help := range topic.Helps {
			if strings.EqualFold(input, help.Name) {
				showHelpItem(player, help)
				return
			}
			for _, keyword := range help.Keywords {
				if strings.EqualFold(input, keyword) {
					showHelpItem(player, help)
					return
				}
			}
		}
	}

	if input != "" {
		player.send("Sorry, I didn't find a help page for that.")
	}
	if len(helpKeywords) > 0 {
		player.send("Help topics: <command name>, commands, emoji, more-emoji, OLC, %v", strings.Join(helpKeywords, ", "))
	} else {
		player.send("No help topics found?")
	}
}

func showHelpItem(player *characterData, help helpData) {
	buf := fmt.Sprintf("HELP: %v (%v)\r\n\r\n%v", help.Name, strings.Join(help.Keywords, ", "), help.Text)
	player.send(buf)
}

var helpKeywords []string

func loadHelps() {
	helpKeywords = []string{}

	dir, err := os.ReadDir(DATA_DIR + HELPS_DIR)
	if err != nil {
		critLog("loadHelps: Unable to read helps dir:  %v", err)
		return
	}
	for _, item := range dir {
		if !item.IsDir() {
			if strings.HasSuffix(item.Name(), ".json") {
				help := loadHelp(item.Name())
				if help != nil {
					//mudLog("Loaded help: %v", item.Name())
					helpFiles = append(helpFiles, help)
					for _, h := range help.Helps {
						helpKeywords = append(helpKeywords, h.Name)
					}
				}
			}
		}
	}

	if len(helpFiles) == 0 {
		critLog("loadHelps: No help files were loaded!")
	}
}

func loadHelp(file string) *helpTopicData {
	data, err := readFile(DATA_DIR + HELPS_DIR + file)

	if err != nil {
		critLog("loadHelp: Unable to read file: %v", err)
		return nil
	}

	newHelpTopic := &helpTopicData{}
	err = json.Unmarshal(data, newHelpTopic)
	if err != nil {
		critLog("loadHelp: Unable to unmarshal the data: %v", err)
		return nil
	}

	for _, help := range newHelpTopic.Helps {
		help.topic = newHelpTopic
	}
	return newHelpTopic
}

func createNewHelpTopic(topic string) {
	newHelpTopic := &helpTopicData{Topic: topic}
	helpFiles = append(helpFiles, newHelpTopic)
}

func createNewHelp(player *characterData, topicStr, title string) {
	for _, topic := range helpFiles {
		if strings.EqualFold(topic.Topic, strings.TrimSpace(topicStr)) {
			newHelp := helpData{topic: topic, Created: time.Now().UTC(),
				Modified: time.Now().UTC(), Authors: []string{player.Name},
				Text: "Work in progress.", Name: title}
			topic.Helps = append(topic.Helps, newHelp)
		}
	}
}

func saveHelps() {
	helpKeywords = []string{}

	for _, topic := range helpFiles {
		if topic.dirty {
			if saveHelp(topic) {
				critLog("--> Saved help file: %v", topic.Topic)
			} else {
				critLog("--> Saving help file failed: %v", topic.Topic)
			}
		}
		for _, h := range topic.Helps {
			helpKeywords = append(helpKeywords, h.Name)
		}
	}
}

// Returns true on save
func saveHelp(helpFile *helpTopicData) bool {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	err := enc.Encode(&helpFile)
	if err != nil {
		critLog("saveHelp: enc.Encode: %v", err.Error())
		return false
	}

	err = saveFile(DATA_DIR+HELPS_DIR+helpFile.Topic+".json", outbuf.Bytes())
	if err != nil {
		critLog("saveHelp: saveFile failed %v", err.Error())
		return false
	}
	return true
}

func makeTestHelp() {
	helpFiles = append(helpFiles, testHelp...)
}

var testHelp []*helpTopicData = []*helpTopicData{
	{
		Topic:    "test",
		Created:  time.Now(),
		Modified: time.Now(),
		Helps: []helpData{
			{
				Name:        "test",
				Created:     time.Now(),
				Modified:    time.Now(),
				Authors:     []string{"System"},
				Keywords:    []string{"test"},
				Description: "This is a test help file.",
				Text:        "This is a test...",
			},
		},
		dirty: true,
	},
}
