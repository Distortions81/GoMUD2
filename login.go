package main

import (
	"strings"
	"time"

	"github.com/martinhoefling/goxkcdpwgen/xkcdpwgen"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"
)

const (
	MAX_PASSPHRASE_LENGTH = 72
	MIN_PASSPHRASE_LENGTH = 8
	PASSPHRASE_HASH_COST  = 10
	MIN_PASS_ENTROPY_BITS = 52
	MAX_CHAR_SLOTS        = 15

	NUM_PASS_SUGGEST = 10

	MAX_LOGIN_LEN = 48
	MIN_LOGIN_LEN = 3

	MAX_NAME_LEN = 30
	MIN_NAME_LEN = 2
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

	CON_CHAR_LIST
	CON_CHAR_CREATE
	CON_CHAR_CREATE_CONFIRM

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

var accountIndex = make(map[string]*accountIndexData)

type accountIndexData struct {
	Login       string
	Fingerprint string
	Added       time.Time
}

type loginStates struct {
	prompt   string
	goPrompt func(desc *descData)
	goDo     func(desc *descData, input string)
	anyKey   bool
	hideInfo bool
}

// These can be defined out of order, neato!
var loginStateList = [CON_MAX]loginStates{
	//Normal login
	CON_LOGIN: {
		prompt: "To create a new account type: NEW\r\nLogin: ",
		goDo:   gLogin,
	},
	CON_PASS: {
		prompt:   "Passphrase: ",
		goDo:     gPass,
		hideInfo: true,
	},
	CON_NEWS: {
		goPrompt: gShowNews,
		goDo:     gNews,
		anyKey:   true,
	},

	//New login
	CON_NEW_LOGIN: {
		prompt: "(LOGIN NAME -- NOT character name. Up to 48 chars long. Spaces allowed.)\r\nNEW login: ",
		goDo:   gNewLogin,
	},
	CON_NEW_LOGIN_CONFIRM: {
		prompt: "(leave blank to choose a new login)\r\nConfirm login: ",
		goDo:   gNewLoginConfirm,
		anyKey: true,
	},
	CON_NEW_PASSPHRASE: {
		goPrompt: gNewPassPrompt,
		goDo:     gNewPassphrase,
		hideInfo: true,
	},
	CON_NEW_PASSPHRASE_CONFIRM: {
		prompt:   "(leave blank to choose a new passphrase)\r\nConfirm passphrase: ",
		goDo:     gNewPassphraseConfirm,
		anyKey:   true,
		hideInfo: true,
	},

	//Character menu
	CON_CHAR_LIST: {
		goPrompt: gCharList,
		goDo:     gCharSelect,
	},
	CON_RECONNECT_CONFIRM: {
		prompt: "That character is already playing!\r\nIf you join, the other connection will be kicked!\r\nAre you sure you want to continue? (y/n)",
		goDo:   gReconnectConfirm,
	},
	CON_CHAR_CREATE: {
		prompt: "Character name:",
		goDo:   gCharNewName,
	},
	CON_CHAR_CREATE_CONFIRM: {
		prompt: "(leave blank to choose a new name)\r\nConfirm character name:",
		goDo:   gCharConfirmName,
		anyKey: true,
	},

	CON_PLAYING: {
		goPrompt: cmdListCmds,
	},
}

// Normal login
func gLogin(desc *descData, input string) {
	if accountIndex[input] != nil {
		err := desc.loadAccount(accountIndex[input].Fingerprint)
		if desc.account != nil {
			desc.sendln("Welcome back %v!", input)
			desc.state = CON_PASS
		} else {
			desc.send(warnBuf)
			desc.sendln("ERROR: Sorry, unable to load that account!")
			critLog("gLogin: %v: %v: Unable to load account: %v (%v)", desc.id, desc.cAddr, input, err)
			desc.close(true)
			return
		}

	} else if strings.EqualFold("new", input) {
		critLog("#%v: %v is creating a new login.", desc.id, desc.cAddr)
		desc.state = CON_NEW_LOGIN

	} else {
		desc.sendln("Invalid login.")
		critLog("#%v: %v tried a login that does not exist!", desc.id, desc.cAddr)
		desc.close(true)
		return
	}
}

func gPass(desc *descData, input string) {

	if bcrypt.CompareHashAndPassword(desc.account.PassHash, []byte(input)) == nil {
		desc.sendln("Passphrase accepted.")
		desc.state = CON_CHAR_LIST
	} else {
		desc.sendln("Incorrect passphrase.")
		critLog("#%v: %v tried a invalid password!", desc.id, desc.cAddr)
		desc.close(true)
	}
}

func gNews(desc *descData, input string) {
	desc.state = CON_PLAYING
}

func gShowNews(desc *descData) {
	desc.sendln("\r\n" + textFiles["news"] + "\r\n[Press return to enter the world]")
}

// New login
func gNewLogin(desc *descData, input string) {
	if nameBad(input) {
		desc.sendln("Sorry, that login is not appropriate.")
		return
	}

	inputLen := len([]byte(input))
	if inputLen >= MIN_LOGIN_LEN && inputLen <= MAX_LOGIN_LEN {
		desc.sendln("Okay, login is: %v", input)
		desc.account = &accountData{
			Login:       input,
			Fingerprint: makeFingerprintString(),
			CreDate:     time.Now(),
			ModDate:     time.Now(),
		}
		desc.state = CON_NEW_LOGIN_CONFIRM
	} else {
		desc.sendln("Sorry, that is not an acceptable login.")
	}
}

func gNewLoginConfirm(desc *descData, input string) {
	if input == desc.account.Login {
		desc.sendln("Okay, login confirmed: %v", input)
		desc.state = CON_NEW_PASSPHRASE
	} else {
		if input == "" {
			desc.sendln("Okay, let's try again.")
			desc.state = CON_NEW_LOGIN
		} else {
			desc.sendln("Login didn't match. Try again, or leave blank to start over.")
		}
	}
}

func gNewPassphrase(desc *descData, input string) {
	//min/max password len
	if len([]byte(input)) > MAX_PASSPHRASE_LENGTH {
		desc.sendln("Sorry, that passphrase is TOO LONG! Try again!")
		return
	} else if len([]byte(input)) < MIN_PASSPHRASE_LENGTH {
		desc.sendln("Sorry, that passphrase is TOO SHORT!")
		return
	}

	//Check if password is decent
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

	desc.account.tempString = input
	desc.state = CON_NEW_PASSPHRASE_CONFIRM
}

func gNewPassphraseConfirm(desc *descData, input string) {
	if input == "" {
		desc.state = CON_NEW_PASSPHRASE
		desc.account.tempString = ""
		desc.sendln("Okay, lets start over!")
		return
	}
	if input == desc.account.tempString {

		desc.sendln("Hashing password... one moment please!")
		var err error
		desc.account.PassHash, err = bcrypt.GenerateFromPassword([]byte(input), PASSPHRASE_HASH_COST)
		desc.sendln("Okay, passwords match!")
		desc.account.tempString = ""

		if err != nil {
			desc.send(warnBuf)
			critLog("ERROR: #%v password hashing failed!!!: %v", desc.id, err.Error())
			desc.sendln("ERROR: something went wrong... Sorry!")
			desc.close(true)
			return
		}
	} else {
		desc.sendln("Passwords did not match! Goodbye!")
		return
	}

	err := desc.account.createAccountDir()
	if err != nil {
		desc.send(warnBuf)
		desc.sendln("Unable to create account! Pleaselet moderators knows!")
		desc.close(true)
		return
	}

	notSaved := desc.account.saveAccount()
	if notSaved {
		desc.send(warnBuf)
		desc.sendln("Unable to save account! Please let moderators know!")
		desc.close(true)
		return
	} else {
		desc.sendln("Account created and saved.")
		newAcc := &accountIndexData{
			Login:       desc.account.Login,
			Fingerprint: desc.account.Fingerprint,
			Added:       time.Now(),
		}
		accountIndex[desc.account.Login] = newAcc
		saveAccountIndex()
	}
	desc.state = CON_CHAR_LIST
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

	buf := "\r\nSuggested passphrases:\n\r\n\r"
	for _, item := range passSuggestions {
		buf = buf + item + "\n\r"
	}
	desc.sendln(buf)
}

func gNewPassPrompt(desc *descData) {
	desc.suggestPasswords()
	desc.sendln("\r\n(minumum 8 characters long)\r\nPassphrase: ")
}
