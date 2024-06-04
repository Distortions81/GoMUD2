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

func reportPanic(format string, args ...interface{}) {
	if r := recover(); r != nil {

		//critLog("Writing '%v' file.", panicDumpName)
		panicFile := fmt.Sprintf("%v/%v/%v-%v", DATA_DIR, PANIC_DIR, time.Now().UnixNano(), panicDumpName)
		f, err := os.Create(panicFile)
		if err == nil {
			debug.WriteHeapDump(f.Fd())
			f.Close()
		} else {
			critLog("Failed to write '%v' file.", panicDumpName)
		}

		_, filename, line, _ := runtime.Caller(4)
		input := fmt.Sprintf(format, args...)
		buf := fmt.Sprintf("(GAME CRASH)\nBUILD:%v-%v-%v\nLabel:%v File: %v Line: %v\nError:%v\n\nStack Trace:\n%v\n", VERSION, VWHEN, CODENAME, input, filepath.Base(filename), line, r, string(debug.Stack()))

		panicLogFile := fmt.Sprintf("%v/%v/%v-%v", DATA_DIR, PANIC_DIR, time.Now().UnixNano(), pnaicLogName)
		os.WriteFile(panicLogFile, []byte(buf), 0660)

		critLog(buf)
	}
}
