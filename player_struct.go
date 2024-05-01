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
	name string
	desc *descData `json:"-"`

	loginTime    time.Time `json:"-"`
	lastOnline   time.Time
	creationDate time.Time
	valid        bool `json:"-"`
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
	id uint64

	login    string
	passHash []byte

	tempPass     string `json:"-"`
	tempCharName string `json:"-"`

	creationDate time.Time
	modDate      time.Time
	lastOnline   time.Time

	characters []string
	banned     banData `json:",omitempty"`
}

type banData struct {
	reason  string `json:",omitempty"`
	date    time.Time
	banBy   string
	revoked bool
}
