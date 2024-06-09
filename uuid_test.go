package main

import (
	"log"
	"testing"
)

func Test(t *testing.T) {

	loadMudID()

	id := makeUUID()
	if !id.hasUUID() {
		log.Fatalln("Failed to generate valid UUID.")
	}

	idStr := id.toString()
	if idStr == "" {
		log.Fatalln("Failed to convert UUID to string.")
	}

	idStrToID := DecodeUUIDString(idStr)
	if id != idStrToID {
		log.Fatalln("UUID string to id failed.")
	}

	var lastUUID uuidData = makeUUID()
	for x := 0; x < 100000000; x++ {
		id := makeUUID()
		if lastUUID.T == id.T {
			log.Fatalf("Duplicate unixnano on interation %v."+NEWLINE, x)
		}
		if lastUUID.R == id.R {
			log.Fatalf("Duplicate rand on interation %v: rand was: %v"+NEWLINE, x, lastUUID.R)
		}
		lastUUID = id
	}

	idA := makeUUID()
	idB := idA
	if !idA.sameUUID(idB) {
		log.Fatalln("sameUUID() didn't detect a match.")
	}
	idC := makeUUID()
	if idA.sameUUID(idC) &&
		idA.M != idC.M &&
		idA.R != idC.R &&
		idA.T != idC.T {
		log.Fatalln("sameUUID() returned match on non-match.")
	}

	var test uuidData
	if test.hasUUID() {
		log.Fatalln("hasUUID() false positive")
	}
	test = makeUUID()
	if !test.hasUUID() {
		log.Fatalln("hasUUID() false negative")
	}
}
