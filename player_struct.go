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
	OLC_OBJECT
	OLC_MOB
	OLC_RESET
	OLC_AREA

	OLC_MAX
)

type characterData struct {
	Version int
	UUID    string
	desc    *descData

	Name  string
	room  *roomData
	Level int

	Channels Bitmask

	OLCMode   int
	OLCInvert bool

	SaveTime time.Time
	CreDate  time.Time
	idleTime time.Time

	loginTime time.Time

	dirty bool
	valid bool
}

type telnetData struct {
	charset, termType string

	charMap *charmap.Charmap
	options *termSettings

	subType   byte
	subMode   bool
	subData   []byte
	subLength int

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
