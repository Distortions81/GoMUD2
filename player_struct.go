package main

import (
	"bufio"
	"net"
	"sync"
	"time"

	"golang.org/x/text/encoding/charmap"
)

type pLEVEL int

const ()

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

	lastChat        string
	chatRepeatCount int

	numLines   int
	lineBuffer []string

	account *accountData
	player  *playerData

	connectTime time.Time
	valid       bool
}

type playerData struct {
	Version int

	Name string
	desc *descData `json:"-"`

	LoginTime time.Time `json:"-"`
	SaveTime  time.Time
	CreDate   time.Time

	dirty bool `json:"-"`
	valid bool `json:"-"`
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

	tempPass     string `json:"-"`
	tempCharName string `json:"-"`

	CreDate    time.Time
	ModDate    time.Time
	LastOnline time.Time

	Characters []string
	Banned     *banData `json:",omitempty"`

	dirty bool `json:"-"`
	valid bool `json:"-"`
}

type banData struct {
	reason  string `json:",omitempty"`
	date    time.Time
	banBy   string
	revoked bool
}
