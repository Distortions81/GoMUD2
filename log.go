package main

import (
	"fmt"
	"log"
	"os"
)

var (
	mlog *os.File
	elog *os.File
)

func startLogs() {
	var err error

	logName := fmt.Sprintf("log/mud-%v-%v-%v.log", bootTime.Day(), bootTime.Month(), bootTime.Year())
	mlog, err = os.OpenFile(logName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Unable to create log file: %v", err.Error())
		os.Exit(1)
		return
	}

	logName = fmt.Sprintf("log/err-%v-%v-%v.log", bootTime.Day(), bootTime.Month(), bootTime.Year())
	elog, err = os.OpenFile(logName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Unable to create log file: %v", err.Error())
		os.Exit(1)
		return
	}

	log.SetOutput(elog)
}

// Log errors, sprintf format
func errLog(format string, args ...any) {
	doLog(elog, format, args...)
}

// Log errors, sprintf format
func critLog(format string, args ...any) {
	doLog(elog, format, args...)

	descLock.Lock()
	defer descLock.Unlock()
	for _, d := range descList {
		d.send("ERROR: "+format, args...)
	}
}

// Log info, sprintf format
func mudLog(format string, args ...any) {
	doLog(mlog, format, args...)
}

func doLog(dest *os.File, format string, args ...any) {
	if args != nil {
		buf := fmt.Sprintf(format, args...)
		if dest != nil {
			dest.WriteString(buf + "\n")
		}
		fmt.Println(buf)
	} else {
		if dest != nil {
			dest.WriteString(format + "\n")
		}
		fmt.Println(format)
	}
}

func closeLogs() {
	if mlog != nil {
		mlog.Close()
	}
	if elog != nil {
		elog.Close()
	}
}
