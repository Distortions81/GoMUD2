package main

import (
	"os"
	"strings"
)

var textFiles map[string]string
var greetBuf string

const textExt = ".txt"

func readTextFiles() {
	textFiles = make(map[string]string)

	dir, err := os.ReadDir(DATA_DIR + TEXTS_DIR)
	if err != nil {
		errLog("readTextFiles: Unable to read texts dir.")
		os.Exit(1)
	}

	for _, fname := range dir {
		if fname.IsDir() {
			continue
		}
		if !strings.HasSuffix(fname.Name(), textExt) {
			continue
		}
		data, err := os.ReadFile(DATA_DIR + TEXTS_DIR + fname.Name())
		if err != nil {
			errLog("readTextFiles: Unable to read file: %v. Error: %v", fname, err.Error())
			os.Exit(1)
		}

		shortName := strings.TrimSuffix(fname.Name(), textExt)
		textFiles[shortName] = ANSIColor(string(data))
		//errLog("readTextFiles: Read: %v", fname.Name())
	}

	//Login prompt
	lItem := loginStateList[CON_LOGIN]
	promptStr := lItem.prompt
	greetBuf = LICENSE + textFiles["greet"] + promptStr
}
