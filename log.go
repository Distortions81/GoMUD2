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
func errLog(format string, args ...interface{}) {
	if args != nil {
		buf := fmt.Sprintf(format, args...)
		if elog != nil {
			elog.WriteString(buf)
		}
		fmt.Println(buf)
	} else {
		if elog != nil {
			elog.WriteString(format)
		}
		fmt.Println(format)
	}
}

// Log info, sprintf format
func mudLog(format string, args ...interface{}) {
	if args != nil {
		buf := fmt.Sprintf(format, args...)
		if mlog != nil {
			mlog.WriteString(buf)
		}
		fmt.Println(buf)
	} else {
		if mlog != nil {
			mlog.WriteString(format)
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
