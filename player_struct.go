package main

import (
	"bufio"
	"net"
	"sync"
	"time"

	"golang.org/x/text/encoding/charmap"
)

type pLEVEL int

const (
	CHAT_SPAM_HISTORY = 10
)

var (
	topID    uint64
	descList []*descData
	descLock sync.Mutex
	playList []*playerData
)

type descData struct {
	id                uint64
	conn              net.Conn
	reader            *bufio.Reader
	state             int
	host, addr, cAddr string

	tls bool

	telnet telnetData

	inputLock      sync.Mutex
	inputBufferLen int
	inputBuffer    []byte

	numLines   int
	lineBuffer []string

	account *accountData
	player  *playerData

	connectTime time.Time
	valid       bool
}

type playerData struct {
	Version     int
	Fingerprint string

	Name string
	desc *descData

	LoginTime time.Time
	SaveTime  time.Time
	CreDate   time.Time

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
	Version     int
	Fingerprint string

	Login    string
	PassHash []byte

	tempPass     string
	tempCharName string

	CreDate    time.Time
	ModDate    time.Time
	LastOnline time.Time

	Characters []accountIndexData
	Banned     *banData `json:",omitempty"`

	dirty bool
	valid bool
}

type banData struct {
	Reason  string `json:",omitempty"`
	Date    time.Time
	BanBy   string
	Revoked bool
}
