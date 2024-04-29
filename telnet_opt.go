package main

import (
	"fmt"

	"golang.org/x/text/encoding/charmap"
)

type termSettings struct {
	ANSI256, ANSI24, UTF, SUPGA bool
	CharMap                     *charmap.Charmap
}

var termTypeMap map[string]*termSettings = map[string]*termSettings{
	//UTF-8 clients
	"AMUDCLIENT": {ANSI256: true, ANSI24: true, UTF: true},
	"ATLANTIS":   {ANSI256: true, UTF: true},
	"BEIP":       {ANSI256: true, UTF: true},
	//ggmud normalizes accents away, always sends UTF-8, but auto-detects recieved?
	//wont/dont sup_gohead & charset, terminal_type bug (reconnects as "hardcopy", "unknown"?),
	"GGMUD": {ANSI256: false, UTF: true},
	"KBTIN": {ANSI256: true, UTF: true},
	//mudlet wont/dont sup_gohead, do charset
	"MUDLET": {ANSI256: true, UTF: true},
	//mudmagic Accepts but does not send UTF-8, and does not accept latin1
	"MUDMAGIC":      {ANSI256: false, UTF: true},
	"MUSHCLIENT":    {ANSI256: true, UTF: true},
	"POTATO":        {ANSI256: true, UTF: true},
	"POWWOW":        {ANSI256: true, UTF: true},
	"TINTIN":        {ANSI256: true, ANSI24: true, UTF: true},
	"TORTILLA":      {ANSI256: true, UTF: true},
	"XTERM256COLOR": {ANSI256: true, UTF: true},

	"BIOMUD":    {ANSI256: true},
	"BLOWTORCH": {ANSI256: true},
	//cmud dont charset, will sup_goahead
	"CMUD":       {ANSI256: true},
	"GMUD":       {ANSI256: false},
	"GNOMEMUD":   {ANSI256: true},
	"JAMOCHAMUD": {ANSI256: false},
	//kild NO RESPONSE: will/do sup_gohead & charset
	"KILDCLIENT": {ANSI256: true},
	"LYNTIN":     {ANSI256: true},
	"KMUDDY":     {ANSI256: false},
	"MCL":        {ANSI256: true},
	"MUBY":       {ANSI256: true},
	"PORTAL":     {ANSI256: false},
	"PUEBLO":     {ANSI256: false},
	"UE":         {ANSI256: false},
	"SIMPLEMU":   {ANSI256: false},
	"SOILED":     {ANSI256: true},
	"TINYFUGUE":  {ANSI256: true},
	"TREBUCHET":  {ANSI256: false},
	"WINTINNET":  {ANSI256: true},
	"ZMUD":       {ANSI256: false},
}

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

var TXT2TermCmd map[string]int

var TermCmd2Txt map[int]string = map[int]string{
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

const (
	SB_IS   = 0
	SB_SEND = 1
)
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
	TXT2TermCmd = make(map[string]int)
	for i, item := range TermCmd2Txt {
		TXT2TermCmd[item] = i
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
