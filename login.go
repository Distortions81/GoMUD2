package main

import (
	"github.com/martinhoefling/goxkcdpwgen/xkcdpwgen"
	passwordvalidator "github.com/wagslane/go-password-validator"
)

const (
	MAX_PASSPHRASE_LENGTH = 128
	PASSPHRASE_HASH_COST  = 10
	MIN_PASS_ENTROPY_BITS = 52
	NUM_PASS_SUGGEST      = 10

	MAX_LOGIN_LEN           = 48
	MIN_PLAYER_LOGIN_LENGTH = 3
)

// Connection states
const (
	CON_DISCONNECTED = iota

	//Greet
	CON_WELCOME
	CON_LOGIN
	CON_PASS
	CON_NEWS

	//New users
	CON_NEW_LOGIN
	CON_NEW_LOGIN_CONFIRM
	CON_NEW_PASSPHRASE
	CON_NEW_PASSPHRASE_CONFIRM
	CON_RECONNECT_CONFIRM

	//Playing
	CON_PLAYING

	/*
	 * Don't delete this
	 * MUST remain at the end
	 * Auto-defines our array size
	 * Never set state to this value
	 */
	CON_MAX
)

type loginStates struct {
	prompt   string
	goPrompt func(desc *descData)
	goDo     func(desc *descData, input string)
	anyKey   bool
}

// These can be defined out of order, neato!
var loginStateList = [CON_MAX]loginStates{
	//Normal login
	CON_LOGIN: {
		prompt: "To create a new account type: NEW.\r\nLogin: ",
		goDo:   gLogin,
	},
	CON_PASS: {
		prompt: "Passphrase: ",
		goDo:   gPass,
	},
	CON_NEWS: {
		goPrompt: gShowNews,
		goDo:     gNews,
		anyKey:   true,
	},

	//New login
	CON_NEW_LOGIN: {
		prompt: "(LOGIN NAME -- NOT character name. Up to 48 chars long. Spaces allowed.)\r\nNew login: ",
		goDo:   gNewLogin,
	},
	CON_NEW_LOGIN_CONFIRM: {
		prompt: "(leave blank to choose a new login).\r\nType login again to confirm: ",
		goDo:   gNewLoginConfirm,
	},
	CON_NEW_PASSPHRASE: {
		goPrompt: gNewPassPrompt,
		goDo:     gNewPassphrase,
	},
	CON_NEW_PASSPHRASE_CONFIRM: {
		prompt: "(leave blank to choose a new passphrase).\r\nType passphrase again to confirm: ",
		goDo:   gNewPassphraseConfirm,
		anyKey: true,
	},
}

// Normal login
func gLogin(desc *descData, input string) {
	if input == "tester" {
		desc.sendln("Welcome %v!", input)
		desc.state = CON_PASS
	} else if input == "new" {
		desc.state = CON_NEW_LOGIN
	} else {
		desc.sendln("Invalid login.")
	}
}

func gPass(desc *descData, input string) {
	if input == "passphrase" {
		desc.send("Pass okay.")
		desc.state = CON_NEWS
	} else {
		desc.send("Incorrect passphrase.")
	}
}

func gNews(desc *descData, input string) {
	//announce arrive here?
	desc.send("Welcome!")
	desc.state = CON_PLAYING
}

func gShowNews(desc *descData) {
	desc.send(textFiles["news"] + "\r\n[Press return to enter the world]")
}

// New login
func gNewLogin(desc *descData, input string) {
	inputLen := len(input)
	if inputLen > 2 && inputLen < MAX_PASSPHRASE_LENGTH {
		desc.sendln("Okay.")
		desc.state = CON_NEW_LOGIN_CONFIRM
	} else {
		desc.sendln("Sorry, that is not an acceptable login.")
	}
}

func gNewLoginConfirm(desc *descData, input string) {
	desc.sendln("Okay, type your login again to confirm")
	desc.state = CON_NEW_PASSPHRASE
}

func gNewPassphrase(desc *descData, input string) {
	ef := passwordvalidator.GetEntropy(input)
	entropy := int(ef)

	if entropy >= 78 {
		desc.sendln("Extremely secure passphrase! %v bits of entropy!", entropy)
	} else if entropy >= 72 {
		desc.sendln("Amazing passphrase! %v bits of entropy!", entropy)
	} else if entropy >= 68 {
		desc.sendln("Great passphrase! %v bits of entropy!", entropy)
	} else if entropy >= 62 {
		desc.sendln("Reasonable passphrase. %v bits of entropy.", entropy)
	} else if entropy >= MIN_PASS_ENTROPY_BITS {
		desc.sendln("Somewhat simple password. %v bits of entropy.", entropy)
	} else {
		desc.sendln("Sorry, please use a more complex passphrase.\r\n%v bits of entropy is NOT good enough.", entropy)
		return
	}
	desc.state = CON_NEW_PASSPHRASE_CONFIRM
}

func gNewPassphraseConfirm(desc *descData, input string) {
	desc.sendln("Okay, type your passphrase again to confirm")
	desc.state = CON_NEWS
}

func (desc *descData) suggestPasswords() {
	var passSuggestions []string
	for x := 0; x < NUM_PASS_SUGGEST; x++ {
		g := xkcdpwgen.NewGenerator()
		g.SetNumWords(3)
		g.SetCapitalize(false)
		sugPass := g.GeneratePasswordString()
		passSuggestions = append(passSuggestions, sugPass)
	}

	buf := "Suggested passphrases:\n\r\n\r"
	for _, item := range passSuggestions {
		buf = buf + item + "\n\r"
	}
	desc.send(buf)
}

func gNewPassPrompt(desc *descData) {
	desc.suggestPasswords()
	desc.send("Passphrase: ")
}
