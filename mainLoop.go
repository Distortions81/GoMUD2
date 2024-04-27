package main

import (
	"time"
)

const (
	ROUND_LENGTH_uS = 250000 //0.25s
)

func mainLoop() {
	var tickNum uint64

	roundTime := time.Duration(ROUND_LENGTH_uS * time.Microsecond)
	for serverState == SERVER_RUNNING {
		tickNum++
		start := time.Now()

		since := roundTime - time.Since(start)
		time.Sleep(since)

		if tickNum%20 == 0 {
			//fmt.Printf("tick %v: slept for: %v\n", tickNum, since.Round(time.Millisecond))
		}
	}
}
