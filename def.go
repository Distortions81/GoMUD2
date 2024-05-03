package main

import "strings"

const (
	VERSION  = "v0.0.4a-05032024-0539"
	CODENAME = "Stellarax"

	LICENSE = "goMUD2: " + VERSION + CODENAME + "\n" +
		"Copyright 2024 Carl Frank Otto III (carlotto81@gmail.com). All rights reserved.\n"

		//Directories
	DATA_DIR    = "data/"
	ACCOUNT_DIR = "accounts/"
	AREA_DIR    = "areas/"
	TEXTS_DIR   = "texts/"
	LOGS_DIR    = "log/"

	//Files
	HELP_FILE      = "helps.json"
	ACCOUNT_FILE   = "acc.json"
	ACC_INDEX_FILE = "accountIndex.json"

	ACCOUNT_VERSION   = 1
	CHARACTER_VERSION = 1

	DEFAULT_CHARSET = "LATIN1"
)

var DEFAULT_CHARMAP = charsetList[DEFAULT_CHARSET]

var makeDirs = []string{
	DATA_DIR,
	DATA_DIR + ACCOUNT_DIR,
	DATA_DIR + AREA_DIR,
	DATA_DIR + TEXTS_DIR,
	LOGS_DIR}

func init() {
	for i := range commandList {
		nameBlacklist = append(nameBlacklist, strings.ToLower(i))
	}
}

var nameBlacklist = []string{
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
	"immortal",
}
