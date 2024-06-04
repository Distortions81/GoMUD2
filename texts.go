package main

import (
	"fmt"
	"os"
	"strings"
)

var textFiles map[string]string
var greetBuf, greetBufNoSSL, fairwellBuf, warnBuf string

const textExt = ".txt"

func readTextFiles() {
	textFiles = make(map[string]string)

	dir, err := os.ReadDir(DATA_DIR + TEXTS_DIR)
	if err != nil {
		critLog("readTextFiles: Unable to read texts dir.")
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
			critLog("readTextFiles: Unable to read file: %v. Error: %v", fname, err.Error())
			os.Exit(1)
		}

		shortName := strings.TrimSuffix(fname.Name(), textExt)
		textFiles[shortName] = string(data)
		//errLog("readTextFiles: Read: %v", fname.Name())
	}

	sslBuf := fmt.Sprintf("Use port %v for SSL.\r\n", *portTLS)

	//Save greet, fairwell and warning
	greetBuf = LICENSE + string(ANSIColor([]byte(textFiles["greet"]))) + loginStateList[CON_ACCOUNT].prompt
	greetBufNoSSL = LICENSE + string(ANSIColor([]byte(textFiles["greet"]))) + sslBuf + loginStateList[CON_ACCOUNT].prompt
	fairwellBuf = textFiles["fairwell"]
	warnBuf = textFiles["warn"]
}
