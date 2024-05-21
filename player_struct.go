package main

import (
	"bufio"
	"net"
	"sync"
	"time"

	"golang.org/x/text/encoding/charmap"
)

type LEVEL int

const (
	LEVEL_ANY    = -1000
	LEVEL_NEWBIE = 0
	LEVEL_PLAYER = 1

	LEVEL_BUILDER     = LEVEL_IMPLEMENTOR - 30
	LEVEL_MODERATOR   = LEVEL_IMPLEMENTOR - 20
	LEVEL_ADMIN       = LEVEL_IMPLEMENTOR - 10
	LEVEL_IMPLEMENTOR = 1000
)

var levelName map[int]string = map[int]string{
	LEVEL_ANY:         "Anyone",
	LEVEL_NEWBIE:      "Newbie",
	LEVEL_PLAYER:      "Player",
	LEVEL_BUILDER:     "Builder",
	LEVEL_MODERATOR:   "Moderator",
	LEVEL_ADMIN:       "Admin",
	LEVEL_IMPLEMENTOR: "Implementor",
}

var (
	topID    uint64
	descList []*descData
	descLock sync.Mutex
	charList []*characterData
)

type descData struct {
	id                uint64
	conn              net.Conn
	reader            *bufio.Reader
	state             int
	host, addr, cAddr string
	idleTime          time.Time

	tls bool

	telnet telnetData

	inputLock sync.Mutex
	inBufLen  int
	inBuf     []byte
	outBuf    []byte
	haveOut   bool

	numInputLines int
	inputLines    []string

	account   *accountData
	character *characterData

	connectTime time.Time
	valid       bool
}

const (
	OLC_NONE = iota
	OLC_ROOM
	OLC_AREA
	OLC_RESET
	OLC_OBJ
	OLC_MOB

	OLC_MAX
)

type IgnoreData struct {
	Name   string
	UUID   string
	Silent bool
	Added  time.Time
}

type characterData struct {
	Version int
	UUID    string
	desc    *descData

	Name  string
	room  *roomData
	Loc   LocData
	Level int

	Channels Bitmask
	Config   Bitmask

	OLCEditor OLCEditorData `json:",omitempty"`
	Ignores   []IgnoreData  `json:",omitempty"`

	SaveTime time.Time
	CreDate  time.Time
	idleTime time.Time

	Tells []tellData `json:",omitempty"`

	loginTime time.Time

	dirty bool
	valid bool
}

type OLCEditorData struct {
	OLCMode   int  `json:",omitempty"`
	OLCInvert bool `json:",omitempty"`

	RoomEditor   EditorData `json:",omitempty"`
	AreaEditor   EditorData `json:",omitempty"`
	ResetEditor  EditorData `json:",omitempty"`
	ObjectEditor EditorData `json:",omitempty"`
	MobEditor    EditorData `json:",omitempty"`

	EditText []string `json:",omitempty"`
}

type EditorData struct {
	TargetUUID string `json:",omitempty"`
	AreaUUID   string `json:",omitempty"`
}

type tellData struct {
	SenderName string
	SenderUUID string
	Message    string
	Sent       time.Time
}

type telnetData struct {
	Charset, termType string

	charMap *charmap.Charmap
	Options *termSettings

	subSeqType byte
	subSeqMode bool
	subSeqData []byte
	subLength  int

	hideEcho bool
}

type accountData struct {
	Version int
	UUID    string
	Level   int

	Login    string
	PassHash []byte

	tempString string

	CreDate time.Time
	ModDate time.Time

	TelnetSettings *telnetData `json:",omitempty"`

	Characters []accountIndexData
	Banned     *banData `json:",omitempty"`

	dirty bool
}

type banData struct {
	Reason  string `json:",omitempty"`
	Date    time.Time
	BanBy   string
	Revoked bool
}
