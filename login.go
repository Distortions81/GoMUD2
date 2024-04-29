package main

const (
	PASSWORD_HASH_COST     = 10
	MAX_PLAYER_NAME_LENGTH = 32
	MIN_PLAYER_NAME_LENGTH = 2
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
		prompt: "New login (not character name): ",
		goDo:   gNewLogin,
	},
	CON_NEW_LOGIN_CONFIRM: {
		prompt: "Type again to condirm: ",
		goDo:   gNewLoginConfirm,
	},
	CON_NEW_PASSWORD: {
		prompt: "Passphrase",
		goDo:   gNewPassword,
	},
	CON_NEW_PASSWORD_CONFIRM: {
		prompt: "Type again to confirm: ",
		goDo:   gNewPasswordConfirm,
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

// New login
func gNewLogin(desc *descData, input string) {
	desc.sendln("Okay.")
	desc.state = CON_NEW_LOGIN_CONFIRM
}

func gNewLoginConfirm(desc *descData, input string) {
	desc.sendln("Okay, type again to confirm")
	desc.state = CON_NEW_PASSWORD
}

func gNewPassword(desc *descData, input string) {
	desc.sendln("Passphrase")
	desc.state = CON_NEW_PASSWORD_CONFIRM
}

func gNewPasswordConfirm(desc *descData, input string) {
	desc.sendln("Okay, type again to confirm")
	desc.state = CON_NEWS
}

func handleCommands(desc *descData, input string) {
	desc.sendln("Echo: %v", input)
}
