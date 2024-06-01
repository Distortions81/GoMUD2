package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/martinhoefling/goxkcdpwgen/xkcdpwgen"
	"golang.org/x/exp/rand"
)

const (
	//Max bcrypt length
	MAX_PASSPHRASE_LENGTH = 72
	MIN_PASSPHRASE_LENGTH = 8

	MAX_CHAR_SLOTS = 15
	//Number of number-suffixed login names
	//to show if the login is taken
	NUM_LOGIN_SUFFIX = 5
	NUM_PASS_SUGGEST = 10

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
	CON_ACCOUNT
	CON_PASS
	CON_CHECK_PASS
	//CON_CHECK_PASS or below
	//have shorter LOGIN_IDLE
	//idle disconnect timer

	//New users
	CON_NEW_ACCOUNT
	CON_NEW_ACCOUNT_CONFIRM
	CON_NEW_PASSPHRASE
	CON_NEW_PASSPHRASE_CONFIRM
	CON_RECONNECT_CONFIRM
	CON_HASH_WAIT

	CON_CHAR_LIST
	CON_CHAR_CREATE
	CON_CHAR_CREATE_CONFIRM

	CON_OPTIONS
	CON_RETIRE
	CON_RETIRE_CONFIRM
	CON_REROLL
	CON_REROLL_CONFIRM

	CON_CHANGE_PASS_OLD
	CON_CHANGE_PASS_NEW
	CON_CHANGE_PASS_CONFIRM

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

var stateName [CON_MAX]string = [CON_MAX]string{
	CON_DISCONNECTED:           "Disconnected",
	CON_WELCOME:                "Welcome",
	CON_ACCOUNT:                "Account",
	CON_PASS:                   "Pass",
	CON_CHECK_PASS:             "Checking pass",
	CON_NEW_ACCOUNT:            "Create new acc",
	CON_NEW_ACCOUNT_CONFIRM:    "Confirm new acc",
	CON_NEW_PASSPHRASE:         "New acc pass",
	CON_NEW_PASSPHRASE_CONFIRM: "New acc pass confirm",
	CON_RECONNECT_CONFIRM:      "Reconnecting",
	CON_HASH_WAIT:              "Hash wait",
	CON_CHAR_LIST:              "Character list",
	CON_CHAR_CREATE:            "Create new char",
	CON_CHAR_CREATE_CONFIRM:    "Confirm new char",
	CON_OPTIONS:                "Options menu",
	CON_RETIRE:                 "Retire menu",
	CON_RETIRE_CONFIRM:         "Retire confirm",
	CON_REROLL:                 "Reroll menu",
	CON_REROLL_CONFIRM:         "Reroll confirm",
	CON_CHANGE_PASS_OLD:        "Changing password",
	CON_CHANGE_PASS_NEW:        "Changing password",
	CON_CHANGE_PASS_CONFIRM:    "Changing password",
	CON_PLAYING:                "Playing",
}

// Quick login lookup
var accountIndex = make(map[string]*accountIndexData)

type accountIndexData struct {
	Login string
	UUID  uuidData
	Added time.Time
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
	CON_ACCOUNT: {
		prompt: "To create a new account type NEW.\r\nAccount name: ",
		goDo:   gLogin,
	},
	CON_PASS: {
		prompt:   "Passphrase: ",
		goDo:     gPass,
		hideInfo: true,
	},

	//New login
	CON_NEW_ACCOUNT: {
		prompt: "(up to 48 characters, spaces and symbols are accepted)\r\nPlease type your desired account name:",
		goDo:   gNewLogin,
	},
	CON_NEW_ACCOUNT_CONFIRM: {
		prompt: "Confirm new login: ",
		goDo:   gNewLoginConfirm,
		anyKey: true,
	},
	CON_NEW_PASSPHRASE: {
		goPrompt: pNewPassPrompt,
		goDo:     gNewPassphrase,
		hideInfo: true,
	},
	CON_NEW_PASSPHRASE_CONFIRM: {
		prompt:   "Confirm passphrase: ",
		goDo:     gNewPassphraseConfirm,
		anyKey:   true,
		hideInfo: true,
	},

	//Character menu
	CON_CHAR_LIST: {
		goPrompt: pCharList,
		goDo:     gCharSelect,
	},
	CON_RECONNECT_CONFIRM: {
		goPrompt: pAlreadyPlayingWarn,
		goDo:     gReconnectConfirm,
	},
	CON_CHAR_CREATE: {
		prompt: "Type 'CANCEL' to cancel.\r\nA-z only, no spaces, numbers or symbols are allowed.\r\nNew character name:",
		goDo:   gCharNewName,
	},
	CON_CHAR_CREATE_CONFIRM: {
		prompt: "Confirm new character name:",
		goDo:   gCharConfirmName,
		anyKey: true,
	},

	//Options area
	CON_OPTIONS: {
		goPrompt: pOptionsMenu,
		goDo:     gOptionsMenu,
	},
	CON_RETIRE: {
		goPrompt: pRetire,
		goDo:     gRetire,
	},
	CON_RETIRE_CONFIRM: {
		goPrompt: pRetireConfirm,
		goDo:     gRetireConfirm,
	},
	CON_REROLL: {
		goPrompt: pReroll,
		goDo:     gReroll,
	},
	CON_REROLL_CONFIRM: {
		goPrompt: pRerollConfirm,
		goDo:     gRerollConfirm,
	},
	CON_CHANGE_PASS_OLD: {
		prompt: "Current passphrase:",
		goDo:   gOldPass,
	},
	CON_CHANGE_PASS_NEW: {
		prompt: "New passphrase:",
		goDo:   gNewPass,
	},
	CON_CHANGE_PASS_CONFIRM: {
		prompt: "Confirm new passphrase:",
		goDo:   gConfirmNewPass,
	},
}

func pOptionsMenu(desc *descData) {
}
func gOptionsMenu(desc *descData, input string) {
}

func pRetire(desc *descData) {
}
func gRetire(desc *descData, input string) {
}

func pRetireConfirm(desc *descData) {
}
func gRetireConfirm(desc *descData, input string) {
}

func pReroll(desc *descData) {
}
func gReroll(desc *descData, input string) {
}

func pRerollConfirm(desc *descData) {
}
func gRerollConfirm(desc *descData, input string) {
}

func gOldPass(desc *descData, input string) {
	hashLock.Lock()
	defer hashLock.Unlock()
	if hashDepth > HASH_DEPTH_MAX {
		desc.send("Sorry, too many passphrase requests are already in the queue. Please try again later.")
		desc.state = CON_DISCONNECTED
		desc.valid = false
	} else {
		desc.send("Checking your passphrase, please wait.")
		hashDepth++
		hashList = append(hashList, &hashData{id: desc.id, desc: desc, hash: desc.account.PassHash, pass: []byte(input), failed: false, doEncrypt: false, started: time.Now(), changePass: true})
		desc.state = CON_HASH_WAIT
	}
}
func gNewPass(desc *descData, input string) {
	//min/max passphrase len
	passLen := len([]byte(input))
	if passLen < MIN_PASSPHRASE_LENGTH ||
		passLen > MAX_PASSPHRASE_LENGTH {
		desc.sendln("Sorry, that passphrase is either over %v or under %v characters. Please try again.", MAX_PASSPHRASE_LENGTH, MIN_PASSPHRASE_LENGTH)
		return
	}

	desc.account.tempString = input
	desc.state = CON_CHANGE_PASS_CONFIRM
}
func gConfirmNewPass(desc *descData, input string) {
	if input == desc.account.tempString {
		desc.state = CON_HASH_WAIT
		desc.idleTime = time.Now()
		desc.sendln("Processing passphrase, please wait.")

		hashLock.Lock()
		if hashDepth > HASH_DEPTH_MAX {
			desc.send("Sorry, too many passphrase requests are already in the queue. Please try again later.")
			desc.state = CON_DISCONNECTED
			desc.valid = false
			desc.account.tempString = ""
			hashLock.Unlock()
			return
		}
		hashDepth++
		hashList = append(hashList, &hashData{id: desc.id, desc: desc, pass: []byte(desc.account.tempString), hash: []byte{}, failed: false, doEncrypt: true, started: time.Now(), changePass: true})

		hashLock.Unlock()
		desc.account.tempString = ""
	} else {
		desc.state = CON_CHANGE_PASS_OLD
		desc.account.tempString = ""
		desc.sendln("Passphrases did not match!")
	}
}

// Normal login
func gLogin(desc *descData, input string) {

	if strings.EqualFold("new", input) {
		critLog("#%v: %v is creating a new login.", desc.id, desc.ip)
		desc.state = CON_NEW_ACCOUNT

	} else if inputLen := len(input); inputLen < MIN_LOGIN_LEN || inputLen > MAX_LOGIN_LEN {
		desc.close()

	} else if accountIndex[input] != nil {
		err := desc.loadAccount(accountIndex[input].UUID)
		if desc.account != nil {
			desc.state = CON_PASS
		} else {
			desc.send(warnBuf + "\r\n")
			desc.sendln("ERROR: Sorry, unable to load that account!")
			critLog("gLogin: %v: %v: Unable to load account: %v (%v)", desc.id, desc.ip, input, err)
			desc.close()

		}
	} else {
		desc.sendln("Login name not found, creating new account.")
		gNewLogin(desc, input)
		desc.state = CON_NEW_ACCOUNT_CONFIRM
	}
}

func gPass(desc *descData, input string) {

	hashLock.Lock()
	defer hashLock.Unlock()
	if hashDepth > HASH_DEPTH_MAX {
		desc.send("Sorry, too many passphrase requests are already in the queue. Please try again later.")
		desc.state = CON_DISCONNECTED
		desc.valid = false
	} else {
		desc.send("Checking your passphrase, please wait.")
		hashDepth++
		hashList = append(hashList, &hashData{id: desc.id, desc: desc, hash: desc.account.PassHash, pass: []byte(input), failed: false, doEncrypt: false, started: time.Now()})
		desc.state = CON_CHECK_PASS
	}
}

// New login
func gNewLogin(desc *descData, input string) {
	inputLen := len([]byte(input))
	if inputLen < MIN_LOGIN_LEN && inputLen > MAX_LOGIN_LEN {
		desc.sendln("Login names must be between %v and %v characters in length.", MIN_LOGIN_LEN, MAX_LOGIN_LEN)
		return
	}
	if !isAccNameAvail(input) {
		var buf string = "A few quick random variations:\r\n"
		for x := 0; x < NUM_LOGIN_SUFFIX; x++ {
			buf = buf + fmt.Sprintf("%v%v\r\n", input, rand.Intn(999))
		}
		buf = buf + "\r\nSorry, that login name is already in use.\r\nPlease pick another one!"
		desc.send(buf)
		return
	}

	desc.account = &accountData{
		Login:   input,
		UUID:    makeUUID(),
		CreDate: time.Now().UTC(),
		ModDate: time.Now().UTC(),
	}
	desc.state = CON_NEW_ACCOUNT_CONFIRM

}

func gNewLoginConfirm(desc *descData, input string) {
	if input == desc.account.Login {
		//Check again! We don't want a collision
		if !isAccNameAvail(input) {
			desc.send("Sorry, that login name is already in use.")
			return
		}
		desc.state = CON_NEW_PASSPHRASE
	} else {
		desc.sendln("Login names didn't match.")
		desc.state = CON_NEW_ACCOUNT
	}
}

func gNewPassphrase(desc *descData, input string) {
	//min/max passphrase len
	passLen := len([]byte(input))
	if passLen < MIN_PASSPHRASE_LENGTH ||
		passLen > MAX_PASSPHRASE_LENGTH {
		desc.sendln("Sorry, that passphrase is either over %v or under %v characters. Please try again.", MAX_PASSPHRASE_LENGTH, MIN_PASSPHRASE_LENGTH)
		return
	}

	desc.account.tempString = input
	desc.state = CON_NEW_PASSPHRASE_CONFIRM
}

func gNewPassphraseConfirm(desc *descData, input string) {
	if input == desc.account.tempString {
		desc.state = CON_HASH_WAIT
		desc.idleTime = time.Now()
		desc.sendln("Processing passphrase, please wait.")

		hashLock.Lock()
		if hashDepth > HASH_DEPTH_MAX {
			desc.send("Sorry, too many passphrase requests are already in the queue. Please try again later.")
			desc.state = CON_DISCONNECTED
			desc.valid = false
			desc.account.tempString = ""
			hashLock.Unlock()
			return
		}
		hashDepth++
		hashList = append(hashList, &hashData{id: desc.id, desc: desc, pass: []byte(desc.account.tempString), hash: []byte{}, failed: false, doEncrypt: true, started: time.Now()})

		hashLock.Unlock()
		desc.account.tempString = ""
	} else {
		desc.state = CON_NEW_PASSPHRASE
		desc.account.tempString = ""
		desc.sendln("Passphrases did not match!")
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

func pNewPassPrompt(desc *descData) {
	desc.suggestPasswords()
	desc.sendln("\r\n(%v to %v characters long)\r\nPassphrase: ", MIN_PASSPHRASE_LENGTH, MAX_PASSPHRASE_LENGTH)
}
