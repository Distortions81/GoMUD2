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

/*
var helpFiles []*helpTopicData = []*helpTopicData{
	{
		Topic:    "basics",
		Created:  time.Now(),
		Modified: time.Now(),
		Helps: []helpData{
			{
				Name:        "Color",
				Created:     time.Now(),
				Modified:    time.Now(),
				Authors:     []string{"System"},
				Keywords:    []string{"ansi", "color", "colour"},
				Description: "This is a test help file.",
				Text:        "WIP.",
			},
		},
		dirty: true,
	},
}*/

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
	if player.desc != nil && strings.EqualFold("commands", input) {
		player.listCommands()
		return
	}
	for _, item := range cmdListStr {
		if strings.EqualFold(strings.TrimSpace(input), item.name) {
			player.send(item.help)
			return
		}
	}
	count := 0
	buf := ""
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
		player.send("Help topics: commands, %v", strings.Join(helpKeywords, ", "))
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
	//helpFiles = []*helpTopicData{}

	dir, err := os.ReadDir(DATA_DIR + HELPS_DIR)
	if err != nil {
		errLog("loadHelps: Unable to read helps dir:  %v", err)
		return
	}
	for _, item := range dir {
		if !item.IsDir() {
			if strings.HasSuffix(item.Name(), ".json") {
				help := loadHelp(item.Name())
				if help != nil {
					errLog("Loaded help: %v", item.Name())
					helpFiles = append(helpFiles, help)
					helpKeywords = append(helpKeywords, help.Topic)
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
		errLog("loadHelp: Unable to read file: %v", err)
		return nil
	}

	newHelpTopic := &helpTopicData{}
	err = json.Unmarshal(data, newHelpTopic)
	if err != nil {
		errLog("loadHelp: Unable to unmarshal the data: %v", err)
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
			saveHelp(topic)
			critLog("--> Saved help file: %v", topic.Topic)
		}
		helpKeywords = append(helpKeywords, topic.Topic)
	}
}

func saveHelp(helpFile *helpTopicData) bool {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	err := enc.Encode(&helpFile)
	if err != nil {
		critLog("saveHelp: enc.Encode: %v", err.Error())
		return true
	}

	err = saveFile(DATA_DIR+HELPS_DIR+helpFile.Topic+".json", outbuf.Bytes())
	if err != nil {
		critLog("saveHelp: saveFile failed %v", err.Error())
		return true
	}
	return false
}
