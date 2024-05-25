package main

import "strings"

const (
	VERSION  = "v0.0.15a-05-25-2024-0146p"
	CODENAME = "Voidstorm"

	LICENSE = "goMUD2: " + VERSION + "-" + CODENAME + "\n" +
		"Copyright 2024 Carl Frank Otto III (carlotto81@gmail.com). All rights reserved.\n"

		//Directories
	DATA_DIR    = "data/"
	ACCOUNT_DIR = "accounts/"
	AREA_DIR    = "areas/"
	HELPS_DIR   = "helps/"
	TEXTS_DIR   = "texts/"

	LOGS_DIR = "log/"

	//Files
	ACCOUNT_FILE   = "acc.json"
	ACC_INDEX_FILE = "accountIndex.json"
	DISABLES_FILE  = "disabled.json"
	BLOCKED_FILE   = "blocked.json"

	ACCOUNT_VERSION   = 1
	CHARACTER_VERSION = 1
	AREA_VERSION      = 1
	DISABLES_VERSION  = 1
	ROOM_VERSION      = 1

	DEFAULT_CHARSET = "LATIN1"
)

var DEFAULT_CHARMAP = charsetList[DEFAULT_CHARSET]

var makeDirs = []string{
	DATA_DIR,
	DATA_DIR + ACCOUNT_DIR,
	DATA_DIR + AREA_DIR,
	DATA_DIR + TEXTS_DIR,
	LOGS_DIR}

// Add commands to reserved names
func init() {
	for i := range cmdMap {
		reservedNames = append(reservedNames, strings.ToLower(i))
	}
}

var reservedNames = []string{
	"new",
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
	"back",
}
