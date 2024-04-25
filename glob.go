package main

const (
	SERVER_BOOTING = iota
	SERVER_RUNNING
	SERVER_SHUTDOWN
)

var (
	serverState int
)
