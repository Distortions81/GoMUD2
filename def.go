package main

const (
	VERSION = "v0.0.1a 04-25-2024 0402p"

	LICENSE = "goMUD " + VERSION + "\n" +
		"COPYRIGHT 2024 Carl Frank Otto III (carlotto81@gmail.com)\n" +
		"License: Mozilla Public License 2.0\n" +
		"This information must remain unmodified, fully intact and shown to end-users.\n"

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
	DATA_DIR + LOGS_DIR}

/*Connection State*/
const (
	CON_STATE_DISCONNECTED = iota

	CON_STATE_WELCOME
	CON_STATE_PASSWORD

	CON_STATE_NEWS
	CON_STATE_RECONNECT_CONFIRM
	CON_STATE_PLAYING

	// New Users
	CON_STATE_NEW_LOGIN
	CON_STATE_NEW_LOGIN_CONFIRM
	CON_STATE_NEW_PASSWORD
	CON_STATE_NEW_PASSWORD_CONFIRM
)
