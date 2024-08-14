package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
)

var (
	elog, mlog *os.File
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
}

// Log errors, sprintf format
func errLog(format string, args ...any) {
	_, filePath, line, _ := runtime.Caller(1)
	file := path.Base(filePath)
	data := fmt.Sprintf(format, args...)
	buf := fmt.Sprintf("%v:%v: %v", file, line, data)
	doLog(elog, buf)
}

// Log errors, sprintf format
func critLog(format string, args ...any) {
	_, filePath, line, _ := runtime.Caller(1)
	file := path.Base(filePath)
	data := fmt.Sprintf(format, args...)
	buf := fmt.Sprintf("%v:%v: %v", file, line, data)
	doLog(elog, buf)

	for _, target := range charList {
		if target.Level >= LEVEL_IMPLEMENTER {
			target.send(buf)
		}
	}
}

// Log info, sprintf format
func mudLog(format string, args ...any) {
	doLog(mlog, format, args...)
}

func doLog(dest *os.File, format string, args ...any) {
	if args != nil {
		buf := fmt.Sprintf(format, args...)
		dest.Write([]byte(buf + NEWLINE))
		fmt.Println(buf)
	} else {
		dest.Write([]byte(format + NEWLINE))
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
