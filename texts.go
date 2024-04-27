package main

import (
	"os"
	"strings"
)

var TextFiles map[string]string

const textExt = ".txt"

func ReadTextFiles() {
	TextFiles = make(map[string]string)

	dir, err := os.ReadDir(DATA_DIR + TEXTS_DIR)
	if err != nil {
		errLog("ReadTextFiles: Unable to read texts dir.")
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
			errLog("ReadTextFiles: Unable to read file: %v. Error: %v", fname, err.Error())
			os.Exit(1)
		}

		shortName := strings.TrimSuffix(fname.Name(), textExt)
		TextFiles[shortName] = ANSIColor(string(data))
		errLog("ReadTextFiles: Read: %v", fname.Name())
	}
}
