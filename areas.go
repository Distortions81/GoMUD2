package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"time"
)

var areaList map[string]*areaData = make(map[string]*areaData)
var sAreaUUID, sRoomUUID string

func makeSystemArea() {
	sAreaUUID, sRoomUUID = makeUUIDString(), makeUUIDString()

	sysRooms := make(map[string]*roomData)
	sysRooms[sRoomUUID] = &roomData{
		Version: 1, UUID: sRoomUUID, VNUM: 0, Name: "The void", Description: "Nothing here."}
	areaList[sAreaUUID] = &areaData{
		Version: 1, UUID: sAreaUUID, VNUM: 0, Name: "system", Rooms: sysRooms}
}

func saveAllAreas() {
	for _, item := range areaList {
		if !item.saveArea() {
			errLog("saved area: %v", fileSafeName(item.Name))
		}
	}
}

// Returns false on error
func (area *areaData) saveArea() bool {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if area == nil {
		critLog("saveArea: Nil area data.")
		return false
	} else if area.UUID == "" {
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
	data, err := readFile(DATA_DIR + AREA_DIR + name)
	if err != nil {
		return nil
	}

	area := &areaData{}
	err = json.Unmarshal(data, area)
	if err != nil {
		errLog("loadPlayer: Unable to unmarshal the data.")
		return nil
	}

	//Add UUID back, we don't want this in the save twice per room
	for r, room := range area.Rooms {
		room.UUID = r
	}

	//Link default system area
	if sAreaUUID == "" || sRoomUUID == "" {
		if area.VNUM == 0 {
			sAreaUUID = area.UUID
			for _, room := range area.Rooms {
				if room.VNUM == 0 {
					sRoomUUID = room.UUID
					break
				}
			}
		}
	}

	return area
}

func loadAllAreas() {
	dir, err := os.ReadDir(DATA_DIR + AREA_DIR)
	if err != nil {
		critLog("Unable to read %v", DATA_DIR+AREA_DIR)
		return
	}
	count := 0
	for _, item := range dir {
		if item.IsDir() {
			continue
		} else if strings.HasSuffix(item.Name(), ".json") {
			count++
			errLog("loading area: %v", item.Name())
			newArea := loadArea(item.Name())
			areaList[newArea.UUID] = newArea
		}
	}
	errLog("Loaded %v areas.", count)
}