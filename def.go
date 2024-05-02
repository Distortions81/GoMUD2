package main

const (
	VERSION  = "v0.0.3a-04302024-1714533857-"
	CODENAME = "Darkflare"

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

var nameBlacklist = []string{
	"new",
	"admin",
	"moderator",
	"mod",
	"someone",
	"something",
	"no one",
	"nobody",
	"imm",
	"immortal",
	"wiz",
	"ooc",
	"ic",
	"sav",
	"save",
	"fuck",
}
