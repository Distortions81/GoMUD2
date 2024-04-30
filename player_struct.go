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

	connectTime time.Time
	valid       bool
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
	tempPass string
	passHash []byte

	creationDate time.Time
	modDate      time.Time
	characters   []string
	banned       banData
}

type banData struct {
	reason  string
	date    time.Time
	banBy   string
	revoked bool
}
