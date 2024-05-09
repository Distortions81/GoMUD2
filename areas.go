package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"strings"
	"time"
)

// Returns false on error
func (area *areaData) saveArea() bool {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if area == nil {
		critLog("saveArea: Nil area data.")
		return false
	} else if !area.UUID.hasUUID() {
		critLog("saveArea: Area '%v' doesn't have a UUID.", fileSafeName(area.Name))
		return false
	}
	area.Version = AREA_VERSION
	area.ModDate = time.Now().UTC()
	fileName := DATA_DIR + AREA_DIR + fileSafeName(area.Name) + ".json"

	err := enc.Encode(&area)
	if err != nil {
		critLog("saveArea: enc.Encode: %v", err.Error())
		return false
	}

	err = saveFile(fileName, outbuf.Bytes())
	if err != nil {
		critLog("saveArea: saveFile failed %v", err.Error())
		return false
	}
	area.dirty = false
	return false
}

func loadArea(name string) *areaData {
	data, err := readFile(DATA_DIR + AREA_DIR + name + ".json")
	if err != nil {
		return nil
	}

	area := &areaData{}
	err = json.Unmarshal(data, area)
	if err != nil {
		errLog("loadPlayer: Unable to unmarshal the data.")
		return nil
	}
	return area
}

func loadAllAreas() {
	dir, err := os.ReadDir(DATA_DIR + AREA_DIR)
	if err != nil {
		critLog("Unable to read %v", DATA_DIR+AREA_DIR)
		return
	}
	var i int
	var item fs.DirEntry
	for i, item = range dir {
		if item.IsDir() {
			continue
		} else if strings.HasSuffix(item.Name(), ".json") {
			errLog("loading area: %v", item.Name())
			loadArea(item.Name())
		}
	}
	errLog("Loaded %v areas.", i)
}
