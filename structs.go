package main

import (
	"bufio"
	"net"
	"sync"
	"time"
)

var (
	topID    uint64
	descList []*descData
	descLock sync.Mutex
)

type descData struct {
	id     uint64
	conn   net.Conn
	reader *bufio.Reader
	state  int

	telnet telnetData

	inputLock      sync.Mutex
	inputBufferLen int
	inputBuffer    []byte
	numLines       int
	lineBuffer     []string

	born time.Time
}

type telnetData struct {
	ansiColor, goAhead, utf bool
	charset, termType       string

	subType   byte
	subMode   bool
	subData   []byte
	subLength int
}
