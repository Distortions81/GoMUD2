package main

import (
	"net"
	"sync"
)

type descData struct {
	conn  *net.Conn
	state uint8

	telnet telnetData

	inputBufferBytes int
	inputBuffer      []byte

	lineBufferLock sync.Mutex
	lineBuffer     []string
	numLines       int
}

type telnetData struct {
	ansiColor, goAhead bool
	charset, termType  string
}
