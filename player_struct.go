package main

import (
	"bufio"
	"net"
	"sync"
	"time"

	"golang.org/x/text/encoding/charmap"
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

	inputLock      sync.Mutex
	inputBufferLen int
	inputBuffer    []byte

	numLines   int
	lineBuffer []string

	account   *accountData
	character *characterData

	connectTime time.Time
	valid       bool
}

type characterData struct {
	Version     int
	Fingerprint string

	Name string
	desc *descData

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
	Version     int
	Fingerprint string

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
