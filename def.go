package main

const (
	VERSION = "v0.0.2a 04-30-2024 1250"

	LICENSE = "goMUD2 " + VERSION + "\n" +
		"Copyright 2024 Carl Frank Otto III (carlotto81@gmail.com), All rights reserved.\n"

		//Directories and files
	DATA_DIR    = "data/"
	ACCOUNT_DIR = "accounts/"
	AREA_DIR    = "areas/"
	TEXTS_DIR   = "texts/"
	LOGS_DIR    = "log/"

	HELP_FILE = "helps.json"

	DEFAULT_CHARSET = "LATIN1"
)

var DEFAULT_CHARMAP = charsetList[DEFAULT_CHARSET]

var makeDirs = []string{
	DATA_DIR,
	DATA_DIR + ACCOUNT_DIR,
	DATA_DIR + AREA_DIR,
	DATA_DIR + TEXTS_DIR,
	LOGS_DIR}
