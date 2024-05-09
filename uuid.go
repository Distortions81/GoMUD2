package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"
)

const mudUUID = 3452345789345

type uuidData struct {
	t, r, m int64
}

func makeUUID() uuidData {
	return uuidData{t: time.Now().UTC().UnixNano(), r: rand.Int63(), m: mudUUID}
}

func (ida uuidData) sameUUID(idb uuidData) bool {
	return ida.hasUUID() && ida.t == idb.t && ida.r == idb.r && ida.m == idb.m
}

func (id uuidData) hasUUID() bool {
	return id.t != 0 && id.m != 0 //Don't check .r, random
}

func (id uuidData) toString() string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, id.t)
	binary.Write(buf, binary.LittleEndian, id.r)
	binary.Write(buf, binary.LittleEndian, id.m)
	return base64.RawURLEncoding.EncodeToString(buf.Bytes())
}

func stringToUUID(input string) uuidData {

	b, _ := base64.RawURLEncoding.DecodeString(input)
	buf := bytes.NewBuffer(b)

	id := uuidData{}
	binary.Read(buf, binary.LittleEndian, &id.t)
	binary.Read(buf, binary.LittleEndian, &id.r)
	binary.Read(buf, binary.LittleEndian, &id.m)
	return id
}

func test() {

	id := makeUUID()
	idStr := id.toString()

	fmt.Println(id)
	fmt.Println(idStr)
	fmt.Println(stringToUUID(idStr))
	fmt.Println(stringToUUID(idStr).toString())
}
