package main

import (
	"os"
	"strconv"
	"time"

	"golang.org/x/exp/rand"
)

func makeFingerprintString(id string) string {
	p1 := RandStringRunes(32)
	p2 := TimeStringRunes()

	if id == "" {
		return (p1 + "-" + p2)
	} else {
		return (p1 + "-" + p2 + "-" + id)
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890_")
var numRunes = len(letterRunes) - 1

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(numRunes)]
	}
	return string(b)
}

func TimeStringRunes() string {
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
func saveFile(filePath string, data []byte) error {
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
