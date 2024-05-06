package main

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

func directionName(exit *exitData) string {
	return dirToStringColor[exit.Direction]
}

func makeFingerprintString() string {
	p1 := randStringRunes(32)
	p2 := timeStringRunes()

	return (p1 + "-" + p2)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890_")
var numRunes = len(letterRunes) - 1

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(numRunes)]
	}
	return string(b)
}

func timeStringRunes() string {
	un := strconv.FormatInt(time.Now().UnixNano(), 10)
	unLen := len(un)

	b := make([]rune, unLen)
	for i := range b {
		var p2 int64
		for x := 0; x < 7; x++ {
			p1, err := strconv.ParseInt(string(un[i]), 10, 64)
			if err != nil {
				continue
			}
			p2 += p1
		}
		b[i] = letterRunes[p2]
	}
	return string(b)
}

// Saves as a temp file, then renames
var saveFileLock sync.Mutex

func saveFile(filePath string, data []byte) error {
	saveFileLock.Lock()
	defer saveFileLock.Unlock()

	tmpName := filePath + ".tmp"
	err := os.WriteFile(tmpName, data, 0755)
	if err != nil {
		critLog("saveFile: ERROR: failed to write file: %v", err.Error())
	}
	err = os.Rename(tmpName, filePath)
	if err != nil {
		critLog("saveFile: ERROR: failed to rename file: %v", err.Error())
	}
	return err
}

func readFile(filePath string) ([]byte, error) {
	saveFileLock.Lock()
	defer saveFileLock.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		errLog("readFile Unable to load file: %v", filePath)
		return nil, err
	}
	return data, nil
}

// Returns false if name is prohibited
func nameReserved(name string) bool {
	for _, item := range reservedNames {
		if item == name {
			return true
		}
	}

	return false
}

func nameReduce(name string) string {
	name = strings.ToLower(name)
	var newName string
	for _, l := range name {
		if l >= 'a' && l <= 'z' {
			newName = newName + string(l)
		}
	}
	return toTitle(newName)
}

func toTitle(s string) string {
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
