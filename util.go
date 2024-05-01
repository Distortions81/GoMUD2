package main

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/exp/rand"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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
	data, err := os.ReadFile(filePath)
	if err != nil {
		errLog("loadAccount: Unable to read account file.")
		return nil, err
	}
	return data, nil
}

// Returns false if name is prohibited
func nameBad(name string) bool {
	for _, item := range nameBlacklist {
		if wideCheck(name, item) {
			return true
		}
	}

	return false
}

// Returns true if there is a partial match
func wideCheck(input string, target string) bool {

	//normalize unicode
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, input)

	//Remove everything but latin letters
	var squished string
	for _, letter := range result {
		if (letter >= 'A' && letter < 'Z') || (letter >= 'a' && letter <= 'z') {
			squished = squished + string(letter)
		}
	}

	//Caps-insensitive matching
	if strings.EqualFold(squished, target) {
		errLog("wideCheck: MATCH: input: %v, target: %v, normalized: %v, squished: %v", input, target, result, squished)
		return true
	}
	return false
}
