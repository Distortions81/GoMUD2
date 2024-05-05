package main

import (
	"bytes"
	"encoding/json"
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
	Title             string
	Created, Modified time.Time

	Keywords []string
	Authors  []string

	Text  string
	topic *helpTopicData
}

func cmdHelp(player *characterData, input string) {
	for _, item := range cmdListStr {
		if strings.EqualFold(strings.TrimSpace(input), item.name) {
			player.send(item.help)
			return
		}
	}
	for _, topic := range helpFiles {
		if topic.Topic == input {

		}
	}

	player.send("Sorry, I didn't find a help page for that.")
}

func loadHelps() {
	helpFiles = []*helpTopicData{}

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
					helpFiles = append(helpFiles, help)
				}
			}
		}
	}

	if len(helpFiles) == 0 {
		critLog("loadHelps: No help files were loaded!")
	}
}

func loadHelp(file string) *helpTopicData {
	data, err := readFile(DATA_DIR + HELPS_DIR)

	if err != nil {
		errLog("loadHelp: Unable to load account file: %v", err)
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
			newHelp := helpData{topic: topic, Created: time.Now(),
				Modified: time.Now(), Authors: []string{player.Name},
				Text: "Work in progress.", Title: title}
			topic.Helps = append(topic.Helps, newHelp)
		}
	}
}

func saveHelps() {
	for _, topic := range helpFiles {
		if topic.dirty {
			saveHelp(topic)
			critLog("--> Saved help file: %v", topic.Topic)
		}
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
