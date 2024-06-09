package main

import "strings"

const (
	VERSION  = "v0.0.24a"
	VWHEN    = "06-09-2024-1213"
	CODENAME = "Galaxium"

	LICENSE = "GOMUD2: " + VERSION + "-" + VWHEN + "-" + CODENAME + "\n" +
		"Copyright 2024 Carl Frank Otto III (carlotto81@gmail.com). All rights reserved.\n"

		//Directories
	DATA_DIR    = "data/"
	ACCOUNT_DIR = "accounts/"
	AREA_DIR    = "areas/"
	HELPS_DIR   = "helps/"
	TEXTS_DIR   = "texts/"
	NOTES_DIR   = "notes/"
	PANIC_DIR   = "panics"

	LOGS_DIR = "log/"

	//Files
	ACCOUNT_FILE   = "acc.json"
	ACC_INDEX_FILE = "accountIndex.json"
	DISABLES_FILE  = "disabled.json"
	BUGS_FILE      = "bugs.json"
	BLOCKED_FILE   = "blocked.json"
	SETTINGS_FILE  = "settings.json"

	ACCOUNT_VERSION   = 1
	CHARACTER_VERSION = 1
	AREA_VERSION      = 1
	DISABLES_VERSION  = 1
	BUGS_VERSION      = 1
	ROOM_VERSION      = 1
	NOTES_VERSION     = 1
	SETTINGS_VERSION  = 1

	DEFAULT_CHARSET = "LATIN1"
)

var DEFAULT_CHARMAP = charsetList[DEFAULT_CHARSET]

var makeDirs = []string{
	DATA_DIR,
	DATA_DIR + ACCOUNT_DIR,
	DATA_DIR + AREA_DIR,
	DATA_DIR + TEXTS_DIR,
	DATA_DIR + NOTES_DIR,
	DATA_DIR + PANIC_DIR,
	LOGS_DIR}

// Add commands to reserved names
func init() {
	for i := range cmdMap {
		reservedNames = append(reservedNames, strings.ToLower(i))
	}
}

var reservedNames = []string{
	"new",
	"back",
	"cancel",
	"options",

	"admin",
	"moderator",
	"mod",
	"someone",
	"something",
	"unknown",
	"noone",
	"nobody",
	"imm",
	"imp",
	"immortal",
}
