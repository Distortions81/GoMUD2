package main

import "strings"

const (
	VERSION  = "v0.0.70a"
	VWHEN    = "07-22-2024-0544p"
	CODENAME = "Novaflux"

	LICENSE = "GOMUD2: " + VERSION + "-" + VWHEN + "-" + CODENAME + NEWLINE +
		"GoMUD2 © 2024 by Carl Frank Otto III is licensed under Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International. To view a copy of this license, visit https://creativecommons.org/licenses/by-nc-sa/4.0/ GitHub: https://github.com/Distortions81/GoMUD2" + NEWLINE

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
	MUDSTATS_VERSION  = 1

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
