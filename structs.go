package main

import (
	"net"
	"sync"
	"time"
)

type descData struct {
	conn  net.Conn
	state int

	telnet telnetData

	inputBufferLen int
	inputBuffer    []byte

	lineBufferLock sync.Mutex
	lineBuffer     []string
	numLines       int

	born time.Time
}

type telnetData struct {
	ansiColor, goAhead bool
	charset, termType  string

	subType   byte
	subMode   bool
	subData   []byte
	subLength int
}
