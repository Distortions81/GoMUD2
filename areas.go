package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"
)

var areaList map[UUIDData]*areaData = make(map[UUIDData]*areaData)
var sysAreaUUID, sysRoomUUID UUIDData

func makeTestArea() {
	sysAreaUUID, sysRoomUUID = makeUUID(), makeUUID()

	sysRooms := make(map[UUIDData]*roomData)
	sysRooms[sysRoomUUID] = &roomData{
		Version: 1, UUID: sysRoomUUID, VNUM: 0, Name: "The void", Description: "You are floating in a void."}
	areaList[sysAreaUUID] = &areaData{
		Version: 1, UUID: sysAreaUUID, VNUM: 0, Name: "system", Rooms: sysRooms}
}

func saveAllAreas(force bool) {
	for _, item := range areaList {
		if !force && !item.dirty {
			continue
		}
		if item.saveArea() {
			critLog("Saved area: %v", fileSafeName(item.Name))
			item.dirty = false
		}
	}
}

// Returns true on save
var areaSaveLock sync.Mutex

func (area *areaData) saveArea() bool {

	if area == nil {
		critLog("saveArea: Nil area data.")
		return false
	}

	areaSaveLock.Lock()
	defer areaSaveLock.Unlock()
	target := *area

	go func(target areaData) {
		outbuf := new(bytes.Buffer)
		enc := json.NewEncoder(outbuf)
		enc.SetIndent("", "\t")

		if !area.UUID.hasUUID() {
			critLog("saveArea: Area '%v' doesn't have a UUID.", fileSafeName(area.Name))
			return
		}
		area.Version = AREA_VERSION
		area.ModDate = time.Now().UTC()
		fileName := DATA_DIR + AREA_DIR + fileSafeName(area.Name) + ".json"

		err := enc.Encode(&area)
		if err != nil {
			critLog("saveArea: enc.Encode: %v", err.Error())
			return
		}

		err = saveFile(fileName, outbuf.Bytes())
		if err != nil {
			critLog("saveArea: saveFile failed %v", err.Error())
			return
		}
		area.dirty = false
	}(target)

	return true
}

func loadArea(name string) (*areaData, error) {
	data, err := readFile(DATA_DIR + AREA_DIR + name)
	if err != nil {
		return nil, err
	}

	area := &areaData{}
	err = json.Unmarshal(data, area)
	if err != nil {
		critLog("loadArea: Unable to unmarshal the data.")
		return nil, err
	}

	//Add UUID back, we don't want this in the save twice per room
	for r, room := range area.Rooms {
		room.UUID = r
	}

	//Link default system area
	if !sysAreaUUID.hasUUID() || !sysRoomUUID.hasUUID() {
		if area.VNUM == 0 {
			sysAreaUUID = area.UUID
			for _, room := range area.Rooms {
				if room.VNUM == 0 {
					sysRoomUUID = room.UUID
					break
				}
			}
		}
	}

	for r, room := range area.Rooms {
		room.pArea = area
		room.UUID = r
	}
	return area, nil
}

func loadAllAreas() {
	dir, err := os.ReadDir(DATA_DIR + AREA_DIR)
	if err != nil {
		critLog("Unable to read %v", DATA_DIR+AREA_DIR)
		return
	}

	for _, item := range dir {
		if item.IsDir() {
			continue
		} else if strings.HasSuffix(item.Name(), ".json") {
			newArea, err := loadArea(item.Name())
			if err != nil {
				continue
			}
			areaList[newArea.UUID] = newArea
			//mudLog("loaded area: %v", item.Name())
		}
	}

	relinkAreaPointers()
}

func relinkAreaPointers() {
	var areaCount, roomCount, exitCount int
	for _, area := range areaList {
		areaCount++
		for _, room := range area.Rooms {
			room.pArea = area
			roomCount++
			for _, exit := range room.Exits {
				exitCount++
				exit.pRoom = areaList[exit.ToRoom.AreaUUID].Rooms[exit.ToRoom.RoomUUID]

			}
		}
	}
	mudLog("Loaded %v area, %v rooms and %v exits..", areaCount, roomCount, exitCount)
}
