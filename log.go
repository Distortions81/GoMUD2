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

func errLog(format string, args ...string) {
	buf := fmt.Sprintf(format, args)
	elog.WriteString(buf)
	fmt.Println(buf)
}

func mudLog(format string, args ...string) {
	buf := fmt.Sprintf(format, args)
	mlog.WriteString(buf)
	fmt.Println(buf)
}

func closeLogs() {
	if mlog != nil {
		mlog.Close()
	}
	if elog != nil {
		elog.Close()
	}
}
