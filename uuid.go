package main

import (
	"encoding/base64"
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
	return id.t != 0 && id.r != 0 && id.m != 0
}

func (id uuidData) toString() string {
	byteSlice := make([]byte, 8)
	for i := 0; i < 8; i++ {
		byteSlice[i] = byte(id.t >> (i * 8) & 0xFF)
	}
	tStr := base64.RawURLEncoding.EncodeToString(byteSlice)

	for i := 0; i < 8; i++ {
		byteSlice[i] = byte(id.r >> (i * 8) & 0xFF)
	}
	rStr := base64.RawURLEncoding.EncodeToString(byteSlice)
	for i := 0; i < 8; i++ {
		byteSlice[i] = byte(id.m >> (i * 8) & 0xFF)
	}
	mStr := base64.RawURLEncoding.EncodeToString(byteSlice)
	return fmt.Sprintf("%v-%v-%v", tStr, rStr, mStr)
}

func stringToUUID(input string) {

}

func test() {

	id := makeUUID()
	fmt.Println(id.toString())

	start := time.Now().UTC()
	var x int
	for x = 0; x < 10000000; x++ {
		test := makeUUID()
		if test.sameUUID(id) {
			fmt.Println("Meep")
		}
	}
	fmt.Printf("%v UUID gen+check took %v\n", x, time.Since(start))
}
