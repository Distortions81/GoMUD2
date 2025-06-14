package main

import (
	"testing"
)

func TestUUID(t *testing.T) {

	loadMudID()

	id := makeUUID()
	if !id.hasUUID() {
		t.Fatalf("Failed to generate valid UUID.")
	}

	idStr := id.toString()
	if idStr == "" {
		t.Fatalf("Failed to convert UUID to string.")
	}

	idStrToID := DecodeUUIDString(idStr)
	if id != idStrToID {
		t.Fatalf("UUID string to id failed.")
	}

	var lastUUID uuidData = makeUUID()
	for x := 0; x < 1000; x++ {
		id := makeUUID()
		if lastUUID.T == id.T {
			t.Fatalf("Duplicate unixnano on interation %v."+NEWLINE, x)
		}
		if lastUUID.R == id.R {
			t.Fatalf("Duplicate rand on interation %v: rand was: %v"+NEWLINE, x, lastUUID.R)
		}
		lastUUID = id
	}

	idA := makeUUID()
	idB := idA
	if !idA.sameUUID(idB) {
		t.Fatalf("sameUUID() didn't detect a match.")
	}
	idC := makeUUID()
	if idA.sameUUID(idC) &&
		idA.M != idC.M &&
		idA.R != idC.R &&
		idA.T != idC.T {
		t.Fatalf("sameUUID() returned match on non-match.")
	}

	var test uuidData
	if test.hasUUID() {
		t.Fatalf("hasUUID() false positive")
	}
	test = makeUUID()
	if !test.hasUUID() {
		t.Fatalf("hasUUID() false negative")
	}
}
