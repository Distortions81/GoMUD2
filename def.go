package main

const (
	VERSION = "v0.0.1a 04-25-2024 0402p"

	LICENSE = "goMUD " + VERSION + "\n" +
		"COPYRIGHT 2024 Carl Frank Otto III (carlotto81@gmail.com)\n" +
		"License: Mozilla Public License 2.0\n" +
		"This information must remain unmodified, fully intact and shown to end-users.\n"

		// Directories and files
	DATA_DIR   = "data/"
	PLAYER_DIR = "players/"
	AREA_DIR   = "areas/"
	TEXTS_DIR  = "texts"
	LOGS_DIR   = "log/"

	HELP_FILE = "helps.json"
)

var makeDirs = []string{
	DATA_DIR,
	DATA_DIR + PLAYER_DIR,
	DATA_DIR + AREA_DIR,
	DATA_DIR + TEXTS_DIR,
	LOGS_DIR}

// Server states
const (
	SERVER_BOOTING = iota
	SERVER_RUNNING
	SERVER_SHUTDOWN
)

// Connection state
const (
	CON_DISCONNECTED = iota

	CON_WELCOME
	CON_PASSWORD

	CON_NEWS
	CON_RECONNECT_CONFIRM
	CON_PLAYING

	// New users
	CON_NEW_LOGIN
	CON_NEW_LOGIN_CONFIRM
	CON_NEW_PASSWORD
	CON_NEW_PASSWORD_CONFIRM
)
