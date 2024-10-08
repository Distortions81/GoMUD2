package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"
)

/* Handles panics */
const panicDumpName = "panic.dat"
const pnaicLogName = "panic.log"

func reportPanic(desc *descData, format string, args ...interface{}) {
	if r := recover(); r != nil {
		now := time.Now().UTC().UnixNano()
		panicFile := fmt.Sprintf("%v/%v/%v-%v", DATA_DIR, PANIC_DIR, now, panicDumpName)
		f, err := os.Create(panicFile)
		if err == nil {
			debug.WriteHeapDump(f.Fd())
			f.Close()
		} else {
			critLog("Failed to write '%v' file.", panicDumpName)
		}

		_, filename, line, _ := runtime.Caller(4)
		input := fmt.Sprintf(format, args...)
		desc.sendln("Sorry, something went wrong running the %v.", input)
		buf := fmt.Sprintf("(GAME PANIC)"+NEWLINE+"BUILD:%v-%v-%v"+NEWLINE+"Label:%v File: %v Line: %v"+NEWLINE+"Error:%v"+NEWLINE+NEWLINE+"Stack Trace:"+NEWLINE+"%v"+NEWLINE, VERSION, VWHEN, CODENAME, input, filepath.Base(filename), line, r, string(debug.Stack()))

		panicLogFile := fmt.Sprintf("%v/%v/%v-%v", DATA_DIR, PANIC_DIR, now, pnaicLogName)
		os.WriteFile(panicLogFile, []byte(buf), 0660)
		critLog(buf)
	}
}
