package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/martinhoefling/goxkcdpwgen/xkcdpwgen"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"
)

const (
	MAX_PASSPHRASE_LENGTH = 72
	MIN_PASSPHRASE_LENGTH = 8

	PASSPHRASE_HASH_COST  = 12
	MIN_PASS_ENTROPY_BITS = 52

	MAX_CHAR_SLOTS     = 15
	NUM_LOGIN_VARIANTS = 5
	NUM_PASS_SUGGEST   = 10

	MAX_LOGIN_LEN = 48
	MIN_LOGIN_LEN = 4

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

	CON_HASH_WAIT
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
		prompt: "[Login name (not character), up to 48 chars long. Spaces and symbols allowed!]\r\nNEW login: ",
		goDo:   gNewLogin,
	},
	CON_NEW_LOGIN_CONFIRM: {
		prompt: "(blank line to go back)\r\nConfirm login: ",
		goDo:   gNewLoginConfirm,
		anyKey: true,
	},
	CON_NEW_PASSPHRASE: {
		goPrompt: gNewPassPrompt,
		goDo:     gNewPassphrase,
		hideInfo: true,
	},
	CON_NEW_PASSPHRASE_CONFIRM: {
		prompt:   "(blank line to go back)\r\nConfirm passphrase: ",
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
		prompt: "(blank line to go back)\r\nConfirm character name:",
		goDo:   gCharConfirmName,
		anyKey: true,
	},

	CON_PLAYING: {
		goPrompt: cmdListCmds,
	},
}

// Normal login
func gLogin(desc *descData, input string) {

	if strings.EqualFold("new", input) {
		critLog("#%v: %v is creating a new login.", desc.id, desc.cAddr)
		desc.state = CON_NEW_LOGIN

	} else if inputLen := len(input); inputLen < MIN_LOGIN_LEN || inputLen > MAX_LOGIN_LEN {
		desc.close()
		return
	} else if accountIndex[input] != nil {
		err := desc.loadAccount(accountIndex[input].Fingerprint)
		if desc.account != nil {
			desc.sendln("Welcome back %v!", input)
			desc.state = CON_PASS
		} else {
			desc.send(warnBuf)
			desc.sendln("ERROR: Sorry, unable to load that account!")
			critLog("gLogin: %v: %v: Unable to load account: %v (%v)", desc.id, desc.cAddr, input, err)
			desc.close()
			return
		}
	} else {
		desc.sendln("Invalid login.")
		critLog("#%v: %v tried a login that does not exist!", desc.id, desc.cAddr)
		desc.close()
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
		desc.state = CON_DISCONNECTED
		desc.valid = false
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
	if !accountNameAvailable(input) {
		var buf string = "Few quick random number suffixes:\r\n"
		for x := 0; x < NUM_LOGIN_VARIANTS; x++ {
			buf = buf + fmt.Sprintf("%v%v\r\n", input, rand.Intn(999))
		}
		buf = buf + "\r\nSorry, that login is already in use. Please pick another one!"
		desc.send(buf)
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
		if !accountNameAvailable(input) {
			desc.send("Sorry, that login is already in use.")
			return
		}
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
		desc.state = CON_HASH_WAIT
		desc.idleTime = time.Now()
		desc.sendln("Hashing password, one moment please...")
		hashLock.Lock()
		hashList = append(hashList, &toHashData{id: desc.id, desc: desc, pass: []byte(desc.account.tempString), hash: []byte{}, failed: false, started: time.Now()})
		hashLock.Unlock()
		desc.account.tempString = ""
	} else {
		desc.sendln("Passwords did not match!")
	}
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

	buf := "\r\nSuggested passphrases:\r\n\r\n"
	for _, item := range passSuggestions {
		buf = buf + item + "\r\n"
	}
	desc.sendln(buf)
}

func gNewPassPrompt(desc *descData) {
	desc.suggestPasswords()
	desc.sendln("\r\n(minumum 8 characters long)\r\nPassphrase: ")
}
