package main

func (desc *descData) interp() {
	var input string

	desc.inputLock.Lock()
	if desc.numLines == 0 {
		desc.inputLock.Unlock()
		return
	} else {
		//Get oldest line
		input = desc.lineBuffer[0]

		if desc.numLines == 1 {
			//If only one line, reset buffer
			desc.lineBuffer = []string{}
			desc.numLines = 0
		} else {
			//otherwise, delete oldest entry and decrement count
			desc.lineBuffer = desc.lineBuffer[1:]
			desc.numLines--
		}
		desc.inputLock.Unlock()
	}

	//Playing, or disconnected
	if desc.state == CON_PLAYING {
		if input != "" {
			handleCommands(desc, input)
		}
		return
	} else if desc.state == CON_DISCONNECTED {
		desc.sendln(textFiles["aurevoir"])
		return
	}

	//Login screen
	if input == "" && !loginStateList[desc.state].anyKey {
		return
	} else {
		loginStateList[desc.state].goDo(desc, input)
	}

	if loginStateList[desc.state].goPrompt != nil {
		loginStateList[desc.state].goPrompt(desc)
	} else {
		desc.send(loginStateList[desc.state].prompt)
	}

}

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
	CON_NEW_PASSWORD
	CON_NEW_PASSWORD_CONFIRM
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

// These can be defined out of order, neato!
var loginStateList = [CON_MAX]loginStates{
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
}

type loginStates struct {
	prompt   string
	goPrompt func(desc *descData)
	goDo     func(desc *descData, input string)
	anyKey   bool
}

func gLogin(desc *descData, input string) {
	if input == "tester" {
		desc.sendln("Welcome %v!", input)
		desc.state = CON_PASS
	} else {
		desc.sendln("Invalid login.")
	}
}

func gPass(desc *descData, input string) {
	if input == "password" {
		desc.send("Pass okay.")
		desc.state = CON_NEWS
	} else {
		desc.send("Incorrect password.")
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
