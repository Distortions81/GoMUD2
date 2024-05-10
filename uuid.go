package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// Used to help prevent ID collisions
const MudIDFile = DATA_DIR + "mud-id.txt"

var MudID int64

type UUIDIntData struct {
	T, R, M int64
}

func makeUUID() UUIDIntData {
	return UUIDIntData{T: time.Now().UTC().UnixNano(), R: rand.Int63(), M: MudID}
}

func makeUUIDString() string {
	id := UUIDIntData{T: time.Now().UTC().UnixNano(), R: rand.Int63(), M: MudID}
	return id.toString()
}

func (id UUIDIntData) toString() string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, id.T)
	binary.Write(buf, binary.LittleEndian, id.R)
	binary.Write(buf, binary.LittleEndian, id.M)
	return base64.RawURLEncoding.EncodeToString(buf.Bytes())
}

func loadMudID() {
	if _, err := os.Stat(MudIDFile); errors.Is(err, os.ErrNotExist) {
		critLog("%v does not exist, creating new mud id.", MudIDFile)
		writeMudID()
		return
	}
	data, err := os.ReadFile(MudIDFile)
	if err != nil {
		critLog("could not read %v", MudIDFile)
		os.Exit(1)
	} else {
		MudID, err = strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			critLog("Unable to parse data in %v, creating new mud id.", MudIDFile)
			os.Remove(MudIDFile)
			writeMudID()
			return
		}
		if MudID == 0 {
			errLog("Read mud id was 0, generating new one.")
			os.Remove(MudIDFile)
			writeMudID()
			return
		}
		errLog("%v loaded: %v", MudIDFile, MudID)
	}
}

func writeMudID() {
	if MudID == 0 {
		MudID = rand.Int63()
		critLog("Invalid mud id, generating new one: %v", MudID)
	}
	err := saveFile(MudIDFile, []byte(fmt.Sprintf("%v", MudID)))
	if err != nil {
		critLog("Unable to write %v", MudIDFile)
		os.Exit(1)
	}
}

// Not generally needed
func (id UUIDIntData) hasUUID() bool {
	return id.T != 0 && id.M != 0 //Don't check .r, random
}

// Not generally needed
func (ida UUIDIntData) sameUUID(idb UUIDIntData) bool {
	return ida.T == idb.T && ida.R == idb.R && ida.M == idb.M
}

// Not generally needed
func DecodeUUIDString(input string) UUIDIntData {
	b, _ := base64.RawURLEncoding.DecodeString(input)
	buf := bytes.NewBuffer(b)

	id := UUIDIntData{}
	binary.Read(buf, binary.LittleEndian, &id.T)
	binary.Read(buf, binary.LittleEndian, &id.R)
	binary.Read(buf, binary.LittleEndian, &id.M)
	return id
}
