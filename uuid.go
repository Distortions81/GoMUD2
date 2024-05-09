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
const MudIDFIle = DATA_DIR + "mud-id.txt"

var MudID int64

type uuidData struct {
	T, R, M int64
}

func makeUUID() uuidData {
	return uuidData{T: time.Now().UTC().UnixNano(), R: rand.Int63(), M: MudID}
}

func (ida uuidData) sameUUID(idb uuidData) bool {
	return ida.hasUUID() && ida.T == idb.T && ida.R == idb.R && ida.M == idb.M
}

func (id uuidData) hasUUID() bool {
	return id.T != 0 && id.M != 0 //Don't check .r, random
}

func (id uuidData) toString() string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, id.T)
	binary.Write(buf, binary.LittleEndian, id.R)
	binary.Write(buf, binary.LittleEndian, id.M)
	return base64.RawURLEncoding.EncodeToString(buf.Bytes())
}

func stringToUUID(input string) uuidData {
	b, _ := base64.RawURLEncoding.DecodeString(input)
	buf := bytes.NewBuffer(b)

	id := uuidData{}
	binary.Read(buf, binary.LittleEndian, &id.T)
	binary.Read(buf, binary.LittleEndian, &id.R)
	binary.Read(buf, binary.LittleEndian, &id.M)
	return id
}

func loadMudID() {
	if _, err := os.Stat(MudIDFIle); errors.Is(err, os.ErrNotExist) {
		critLog("%v does not exist, creating new mud id.", MudIDFIle)
		writeMudID()
		return
	}
	data, err := os.ReadFile(MudIDFIle)
	if err != nil {
		critLog("could not read %v", MudIDFIle)
		os.Exit(1)
	} else {
		MudID, err = strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			critLog("Unable to parse data in %v, creating new mud id.", MudIDFIle)
			os.Remove(MudIDFIle)
			writeMudID()
			return
		}
		if MudID == 0 {
			errLog("Read mud id was 0, generating new one.")
			os.Remove(MudIDFIle)
			writeMudID()
			return
		}
		errLog("%v loaded: %v", MudIDFIle, MudID)
	}
}

func writeMudID() {
	if MudID == 0 {
		MudID = rand.Int63()
		critLog("Invalid mud id, generating new one: %v", MudID)
	}
	err := os.WriteFile(MudIDFIle, []byte(fmt.Sprintf("%v", MudID)), 0755)
	if err != nil {
		critLog("Unable to write %v", MudIDFIle)
		os.Exit(1)
	}
}
