package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/hako/durafmt"
	"github.com/martinhoefling/goxkcdpwgen/xkcdpwgen"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"
)

const (
	MAX_PASSPHRASE_LENGTH = 72
	MIN_PASSPHRASE_LENGTH = 8

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
	CON_HASH_WAIT

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
	CON_LOGIN: {
		prompt: "To create a new account type: NEW\r\nAccount name: ",
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
		prompt: "(up to 48 characters, spaces and symbols are accepted)\r\nPlease type your desired account name:",
		goDo:   gNewLogin,
	},
	CON_NEW_LOGIN_CONFIRM: {
		prompt: "(type 'back' to go back)\r\nConfirm account name: ",
		goDo:   gNewLoginConfirm,
		anyKey: true,
	},
	CON_NEW_PASSPHRASE: {
		goPrompt: gNewPassPrompt,
		goDo:     gNewPassphrase,
		hideInfo: true,
	},
	CON_NEW_PASSPHRASE_CONFIRM: {
		prompt:   "(type 'back' to go back)\r\nConfirm passphrase: ",
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
		goPrompt: gAlreadyPlayingWarn,
		goDo:     gReconnectConfirm,
	},
	CON_CHAR_CREATE: {
		prompt: "A-z only, no spaces, number or symbols are allowed.\r\nCharacter name:",
		goDo:   gCharNewName,
	},
	CON_CHAR_CREATE_CONFIRM: {
		prompt: "(type 'back' to go back)\r\nConfirm character name:",
		goDo:   gCharConfirmName,
		anyKey: true,
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
		err := desc.loadAccount(accountIndex[input].UUID)
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
		desc.sendln("Login name not found.")
		critLog("#%v: %v tried a login that does not exist!", desc.id, desc.cAddr)
		desc.close()
		return
	}
}

func gPass(desc *descData, input string) {

	if bcrypt.CompareHashAndPassword(desc.account.PassHash, []byte(input)) == nil {
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
	desc.sendln("\r\n" + textFiles["news"] + "\r\n[Press enter to proceed]")
}

// New login
func gNewLogin(desc *descData, input string) {
	inputLen := len([]byte(input))
	if inputLen < MIN_LOGIN_LEN && inputLen > MAX_LOGIN_LEN {
		desc.sendln("Login names must be between %v and %v characters in length.", MIN_LOGIN_LEN, MAX_LOGIN_LEN)
		return
	}
	if !accountNameAvailable(input) {
		var buf string = "A few quick random variations:\r\n"
		for x := 0; x < NUM_LOGIN_VARIANTS; x++ {
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
	desc.state = CON_NEW_LOGIN_CONFIRM

}

func gNewLoginConfirm(desc *descData, input string) {
	if input == desc.account.Login {
		if !accountNameAvailable(input) {
			desc.send("Sorry, that login name is already in use.")
			return
		}
		desc.state = CON_NEW_PASSPHRASE
	} else {
		if input == "" || strings.EqualFold(input, "back") {
			desc.sendln("Okay, let's try again.")
			desc.state = CON_NEW_LOGIN
		} else {
			desc.sendln("Login names didn't match.\r\nTry again, to type 'back' to go back.")
		}
	}
}

func gNewPassphrase(desc *descData, input string) {
	//min/max password len
	passLen := len([]byte(input))
	if passLen < MIN_PASSPHRASE_LENGTH &&
		passLen > MAX_PASSPHRASE_LENGTH {
		desc.sendln("Sorry, that passphrase is TOO LONG! Try again!")
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
		desc.sendln("Please try again with a more complex passphrase.\r\n")
		return
	}

	desc.account.tempString = input
	desc.state = CON_NEW_PASSPHRASE_CONFIRM
}

func gNewPassphraseConfirm(desc *descData, input string) {
	if input == "" || strings.EqualFold(input, "back") {
		desc.state = CON_NEW_PASSPHRASE
		desc.account.tempString = ""
		desc.sendln("Okay, lets try again.")
		return
	}
	if input == desc.account.tempString {
		desc.state = CON_HASH_WAIT
		desc.idleTime = time.Now()
		desc.sendln("Processing password, one moment please...")

		hashLock.Lock()
		hashDepth := len(hashList)
		if hashDepth > HASH_DEPTH_MAX {
			desc.send("Sorry, %v other password requests are already in the queue. Try again later!", hashDepth)
			desc.state = CON_NEW_PASSPHRASE
			desc.account.tempString = ""
			hashLock.Unlock()
			return
		} else {
			if hashDepth > 0 {
				willTake := int(math.Round(lastHashTime.Seconds())) * (hashDepth + 1)
				if willTake > 3 {
					desc.send("%v password requests in the queue. Approx wait time: %v seconds.", hashDepth+1, willTake)
				}
			} else {
				if lastHashTime.Seconds() > 3 {
					desc.send("Should take about %v.", durafmt.Parse(lastHashTime.Round(time.Second)).LimitFirstN(2))
				}
			}
		}
		hashList = append(hashList, &toHashData{id: desc.id, desc: desc, pass: []byte(desc.account.tempString), hash: []byte{}, failed: false, doEncrypt: true, started: time.Now()})

		hashLock.Unlock()
		desc.account.tempString = ""
	} else {
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

func gNewPassPrompt(desc *descData) {
	desc.suggestPasswords()
	desc.sendln("\r\n(%v to %v characters long)\r\nPassphrase: ", MIN_PASSPHRASE_LENGTH, MAX_PASSPHRASE_LENGTH)
}
