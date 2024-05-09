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

	idStrToID := stringToUUID(idStr)
	if id != idStrToID {
		log.Fatalln("UUID string to id failed.")
	}

	var lastUUID uuidData = makeUUID()
	for x := 0; x < 1000000; x++ {
		id := makeUUID()
		if lastUUID.t == id.t {
			log.Fatalf("Duplicate unixnano on interation %v.\n", x)
		}
		if lastUUID.r == id.r {
			log.Fatalf("Duplicate rand on interation %v (false positive is possible): rand: %v\n", x, lastUUID.r)
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
		idA.m != idC.m &&
		idA.r != idC.r &&
		idA.t != idC.t {
		log.Fatalln("sameUUID() returned match on non-match.")
	}
}
