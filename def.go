package main

const (
	VERSION = "v0.0.1a 04-25-2024 0402p"

	LICENSE = "goMUD2 " + VERSION + "\n" +
		"COPYRIGHT 2024 Carl Frank Otto III (carlotto81@gmail.com)\n" +
		"License: Mozilla Public License 2.0\n" +
		"This information must remain unmodified, fully intact and shown to end-users.\n"

		//Directories and files
	DATA_DIR   = "data/"
	PLAYER_DIR = "players/"
	AREA_DIR   = "areas/"
	TEXTS_DIR  = "texts/"
	LOGS_DIR   = "log/"

	HELP_FILE = "helps.json"

	DEFAULT_CHARSET = "LATIN1"
)

var DEFAULT_CHARMAP = charsetList[DEFAULT_CHARSET]

var makeDirs = []string{
	DATA_DIR,
	DATA_DIR + PLAYER_DIR,
	DATA_DIR + AREA_DIR,
	DATA_DIR + TEXTS_DIR,
	LOGS_DIR}
