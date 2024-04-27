package main

import (
	"fmt"
	"time"
)

const (
	ROUND_LENGTH_uS = 250000 //0.25s
)

func mainLoop() {

	roundTime := time.Duration(ROUND_LENGTH_uS * time.Microsecond)
	for serverState == SERVER_RUNNING {
		start := time.Now()

		since := roundTime - time.Since(start)
		time.Sleep(since)
		fmt.Printf("tick: slept for: %v\n", since)
	}
}
