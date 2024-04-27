package main

import (
	"time"
)

const (
	ROUND_LENGTH_uS = 250000 //0.25s
)

var gameTick chan uint64

func mainLoop() {
	var tickNum uint64

	gameTick = make(chan uint64)

	roundTime := time.Duration(ROUND_LENGTH_uS * time.Microsecond)
	for serverState == SERVER_RUNNING {
		tickNum++
		start := time.Now()

		since := roundTime - time.Since(start)
		time.Sleep(since)

		descLock.Lock()
		for _, desc := range descList {
			desc.interp()
		}
		descLock.Unlock()

		gameTick <- tickNum
	}
}
