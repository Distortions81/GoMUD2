package main

import (
	"fmt"

	"golang.org/x/text/encoding/charmap"
)

const (
	MTTS_ANSI = iota
	MTTS_VT100
	MTTS_UTF8
	MTTS_256
	MTTS_MOUSE
	MTTS_OSC_COLOR
	MTTS_SCREEN_READER
	MTTS_PROXY
	MTTS_TRUECOLOR
	MTTS_MNES
	MTTS_MSLP
	MTTS_SSL
)

type MTTSNameData struct {
	Name string
}

var MTTSNames map[int]MTTSNameData = map[int]MTTSNameData{
	MTTS_ANSI:          {Name: "ANSI-Color"},
	MTTS_VT100:         {Name: "VT100"},
	MTTS_UTF8:          {Name: "UTF-8"},
	MTTS_256:           {Name: "256-Colors"},
	MTTS_MOUSE:         {Name: "Mouse-Support"},
	MTTS_OSC_COLOR:     {Name: "OSC Color"},
	MTTS_SCREEN_READER: {Name: "Screen-reader"},
	MTTS_PROXY:         {Name: "Proxy"},
	MTTS_TRUECOLOR:     {Name: "True-Color"},
	MTTS_MNES:          {Name: "MNES"},
	MTTS_MSLP:          {Name: "MSLP"},
	MTTS_SSL:           {Name: "Secure-Socket"},
}

type termSettings struct {
	MTTS                           Bitmask
	SuppressGoAhead, NAWS, HasMTTS bool
	TermWidth, TermHeight          int
	charMap                        *charmap.Charmap
}

var termTypeMap map[string]*termSettings = map[string]*termSettings{
	//amudclient Java, didnt try it
	"AMUDCLIENT": {MTTS: MTTS_ANSI | MTTS_256 | MTTS_TRUECOLOR | MTTS_UTF8},

	//atlantis Macintosh / OS X, didn't test
	"ATLANTIS": {MTTS: MTTS_ANSI | MTTS_256 | MTTS_UTF8},

	//beip Worked fine
	"BEIP": {MTTS: MTTS_ANSI | MTTS_256 | MTTS_UTF8},

	//ggmud normalizes accents away, always sends UTF-8, but auto-detects recieved?
	//terminal_type bug (reconnects as "hardcopy", "unknown"?),
	"GGMUD":    {MTTS: MTTS_ANSI | MTTS_256 | MTTS_UTF8},
	"HARDCOPY": {MTTS: MTTS_ANSI | MTTS_256 | MTTS_UTF8},
	"UNKNOWN":  {MTTS: MTTS_ANSI | MTTS_256 | MTTS_UTF8},
	//ggmud

	//kbtin Didn't find any binaries, just source, did not test
	"KBTIN": {MTTS: MTTS_ANSI | MTTS_256 | MTTS_UTF8},

	//mudlet Works fine
	"MUDLET": {MTTS: MTTS_ANSI | MTTS_256 | MTTS_UTF8},

	//mudmagic will accept but does not send UTF-8, does not accept latin1
	//eats whole lines with accents, crashed more than once, no termtype
	//"MUDMAGIC": {ANSI256: false, UTF: true},

	//mushclient Works fine
	"MUSHCLIENT": {MTTS: MTTS_ANSI | MTTS_256},

	//potato Works fine if you never send GA
	"POTATO": {MTTS: MTTS_ANSI | MTTS_256 | MTTS_UTF8, SuppressGoAhead: true},

	//POWWOW WINDOWS, uses DOS/OEM/CP437
	"CYGWIN": {MTTS: MTTS_ANSI | MTTS_256, charMap: charmap.CodePage437},

	//tintin Newline issues on linux?
	"TINTIN": {MTTS: MTTS_ANSI | MTTS_256 | MTTS_UTF8},

	//I couldn't figure out how to connect with it
	//"TORTILLA": {ANSI256: true, UTF: true},

	//Unable to connect to anything
	//"BIOMUD": {ANSI256: true},

	//blowtorch Didn't test, no android devices
	"BLOWTORCH": {MTTS: MTTS_ANSI | MTTS_256},

	//Works, but eats whole lines with accent characters?
	"CMUD": {MTTS: MTTS_ANSI | MTTS_256},

	//Works, but normalizes text, no termtype
	//"GMUD":       {ANSI256: false},

	//Would not run
	//"GNOMEMUD": {ANSI256: true},

	//Java client, didn't open for me
	//"JAMOCHAMUD": {ANSI256: false},

	//kild works fine
	"KILDCLIENT": {MTTS: MTTS_ANSI | MTTS_256},

	//Didn't run
	//"LYNTIN": {ANSI256: true},

	//No binary found
	//"KMUDDY": {ANSI256: false},

	//No binary found
	//"MCL": {ANSI256: true},

	//No results
	//"MUBY": {ANSI256: true},

	//Wouldn't launch
	//"PORTAL": {ANSI256: false},

	//Works fine but normalizes text
	"PUEBLO": {MTTS: MTTS_ANSI},

	//Works fine, but termtype is just 'ansi'
	//"SIMPLEMU": {ANSI256: false},

	//No binary found
	//"SOILED": {ANSI256: true},

	//No binary found
	//"TINYFUGUE": {ANSI256: true},

	//Didnt open for me
	//"TREBUCHET": {ANSI256: false},

	//Worked fine, no termtype
	//"WINTINNET": {ANSI256: true},

	//Worked fine
	"ZMUD": {MTTS: MTTS_ANSI},

	//Generic terminal
	"XTERM256COLOR":  {MTTS: MTTS_ANSI | MTTS_256 | MTTS_UTF8},
	"XTERMTRUECOLOR": {MTTS: MTTS_ANSI | MTTS_256 | MTTS_TRUECOLOR | MTTS_UTF8},
	"VT100":          {MTTS: MTTS_ANSI | MTTS_VT100},
	"ANSI":           {MTTS: MTTS_ANSI},
	"MONO":           {},
	"DUMB":           {},

	//Someone said MUDRammer supports UTF-8, check?
}

// DO NOT CHANGE ORDER
const (
	TermCmd_SE = iota + 240
	TermCmd_NOP
	TermCmd_DATA_MARK
	TermCmd_BREAK
	TermCmd_INTERRUPT
	TermCmd_ABORT
	TermCmd_ARE_YOU_THERE
	TermCmd_ERASECHAR
	TermCmd_ERASELINE
	TermCmd_GOAHEAD
	TermCmd_SB
	TermCmd_WILL
	TermCmd_WONT
	TermCmd_DO
	TermCmd_DONT
	TermCmd_IAC
)

var txt2TermCmd map[string]int

var termCmd2Txt map[int]string = map[int]string{
	TermCmd_SE:            "SE",
	TermCmd_NOP:           "NOP",
	TermCmd_DATA_MARK:     "DATA_MARK",
	TermCmd_BREAK:         "BREAK",
	TermCmd_INTERRUPT:     "INTERRUPT",
	TermCmd_ABORT:         "ABORT",
	TermCmd_ARE_YOU_THERE: "ARE_YOU_THERE",
	TermCmd_ERASECHAR:     "ERASECHAR",
	TermCmd_ERASELINE:     "ERASELINE",
	TermCmd_GOAHEAD:       "GOAHEAD",
	TermCmd_SB:            "SB",
	TermCmd_WILL:          "WILL",
	TermCmd_WONT:          "WONT",
	TermCmd_DO:            "DO",
	TermCmd_DONT:          "DONT",
	TermCmd_IAC:           "IAC",
}

// DO NOT CHANGE ORDER
const (
	TermOpt_BINARY = iota
	TermOpt_ECHO
	TermOpt_RECONNECTION
	TermOpt_SUP_GOAHEAD
	TermOpt_MESSAGE_SIZE
	TermOpt_STATUS
	TermOpt_TIMING_MARK
	TermOpt_REMOTE_CTE
	TermOpt_LINE_WIDTH
	TermOpt_PAGE_SIZE
	TermOpt_OUT_CRD
	TermOpt_OUT_HTS
	TermOpt_OUT_HTD
	TermOpt_OUT_FFD
	TermOpt_OUT_VTS
	TermOpt_OUT_VTD
	TermOpt_OUT_LFD
	TermOpt_EXTENDED_ASCII
	TermOpt_LOGOUT
	TermOpt_BYTE_MACRO
	TermOpt_DATA_ENTRY_TERMINAL
	TermOpt_SUPDUP
	TermOpt_SUPDUP_OUT
	TermOpt_SEND_LOC
	TermOpt_TERMINAL_TYPE
	TermOpt_END_OF_RECORD
	TermOpt_TACACS
	TermOpt_OUTPUT_MARKING
	TermOpt_TERMINAL_LOC_NUM
	TermOpt_TELNET_3270
	TermOpt_X3_PAD
	TermOpt_WINDOW_SIZE
	TermOpt_TERM_SPEED
	TermOpt_REMOTE_FLOW_CONTROL
	TermOpt_LINEMODE
	TermOpt_DISPLAY_LOC
	TermOpt_ENV_OPT
	TermOpt_AUTH_OPT
	TermOpt_ENC_OPT
	TermOpt_NEW_ENV_OPT
	TermOpt_TN3270E
	TermOpt_XAUTH
	TermOpt_CHARSET
	TermOpt_TELNET_REMTOE_SERIAL
	TermOpt_COM_PORT_CONTROL
	TermOpt_SUP_LOCAL_ECHO
	TermOpt_START_TLS
	TermOpt_KERMIT
	TermOpt_SEND_URL
	TermOpt_FORWARD_X

	TermOpt_MCCP  = 85
	TermOpt_MCCP2 = 86
	TermOpt_MCCP3 = 87

	TermOpt_PRAGMA_LOGON = iota + 88
	TermOpt_SSPI_LOGON
	TermOpt_PRAGMA_HEARTBEAT
	TermOpt_EXTENDED_OPTIONS_LIST = 255
)

var TXT2TermOpt map[string]int

var TermOpt2TXT map[int]string = map[int]string{
	TermOpt_BINARY:               "BINARY",
	TermOpt_ECHO:                 "ECHO",
	TermOpt_RECONNECTION:         "RECONNECTION",
	TermOpt_SUP_GOAHEAD:          "SUP_GOAHEAD",
	TermOpt_MESSAGE_SIZE:         "MESSAGE_SIZE",
	TermOpt_STATUS:               "STATUS",
	TermOpt_TIMING_MARK:          "TIMING_MARK",
	TermOpt_REMOTE_CTE:           "REMOTE_CTE",
	TermOpt_LINE_WIDTH:           "LINE_WIDTH",
	TermOpt_PAGE_SIZE:            "PAGE_SIZE",
	TermOpt_OUT_CRD:              "OUT_CRD",
	TermOpt_OUT_HTS:              "OUT_HTS",
	TermOpt_OUT_HTD:              "OUT_HTD",
	TermOpt_OUT_FFD:              "OUT_FFD",
	TermOpt_OUT_VTS:              "OUT_VTS",
	TermOpt_OUT_VTD:              "OUT_VTD",
	TermOpt_OUT_LFD:              "OUT_LFD",
	TermOpt_EXTENDED_ASCII:       "EXTENDED_ASCII",
	TermOpt_LOGOUT:               "LOGOUT",
	TermOpt_BYTE_MACRO:           "BYTE_MACRO",
	TermOpt_DATA_ENTRY_TERMINAL:  "DATA_ENTRY_TERMINAL",
	TermOpt_SUPDUP:               "SUPDUP",
	TermOpt_SUPDUP_OUT:           "SUPDUP_OUT",
	TermOpt_SEND_LOC:             "SEND_LOC",
	TermOpt_TERMINAL_TYPE:        "TERMINAL_TYPE",
	TermOpt_END_OF_RECORD:        "END_OF_RECORD",
	TermOpt_TACACS:               "TACACS",
	TermOpt_OUTPUT_MARKING:       "OUTPUT_MARKING",
	TermOpt_TERMINAL_LOC_NUM:     "TERMINAL_LOC_NUM",
	TermOpt_TELNET_3270:          "TELNET_3270",
	TermOpt_X3_PAD:               "X3_PAD",
	TermOpt_WINDOW_SIZE:          "WINDOW_SIZE",
	TermOpt_TERM_SPEED:           "TERM_SPEED",
	TermOpt_REMOTE_FLOW_CONTROL:  "REMOTE_FLOW_CONTROL",
	TermOpt_LINEMODE:             "LINEMODE",
	TermOpt_DISPLAY_LOC:          "DISPLAY_LOC",
	TermOpt_ENV_OPT:              "ENV_OPT",
	TermOpt_AUTH_OPT:             "AUTH_OPT",
	TermOpt_ENC_OPT:              "ENC_OPT",
	TermOpt_NEW_ENV_OPT:          "NEW_ENV_OPT",
	TermOpt_TN3270E:              "TN3270E",
	TermOpt_XAUTH:                "XAUTH",
	TermOpt_CHARSET:              "CHARSET",
	TermOpt_TELNET_REMTOE_SERIAL: "TELNET_REMTOE_SERIAL",
	TermOpt_COM_PORT_CONTROL:     "COM_PORT_CONTROL",
	TermOpt_SUP_LOCAL_ECHO:       "SUP_LOCAL_ECHO",
	TermOpt_START_TLS:            "START_TLS",
	TermOpt_KERMIT:               "KERMIT",
	TermOpt_SEND_URL:             "SEND_URL",
	TermOpt_FORWARD_X:            "FORWARD_X",

	TermOpt_PRAGMA_LOGON:          "TermOpt_PRAGMA_LOGON",
	TermOpt_SSPI_LOGON:            "SSPI_LOGON",
	TermOpt_PRAGMA_HEARTBEAT:      "PRAGMA_HEARTBEAT",
	TermOpt_EXTENDED_OPTIONS_LIST: "EXTENDED_OPTIONS_LIST",
}

// DO NOT CHANGE ORDER
const (
	SB_IS   = 0
	SB_SEND = 1
)

// DO NOT CHANGE ORDER
const (
	SB_REQ = iota + 1
	SB_ACCEPTED
	SB_REJECTED
	SB_TTABLE_IS
	SB_TTABLE_REJECTED
	SB_TTABLE_ACK
	SB_TTABLE_NAK
)

func init() {
	txt2TermCmd = make(map[string]int)
	for i, item := range termCmd2Txt {
		txt2TermCmd[item] = i
	}

	for x := 0; x < 255; x++ {
		if TermOpt2TXT[x] == "" {
			TermOpt2TXT[x] = fmt.Sprintf("Unknown: %d", x)
		}
	}

	TXT2TermOpt = make(map[string]int)
	for i, item := range TermOpt2TXT {
		TXT2TermOpt[item] = i
	}

}
