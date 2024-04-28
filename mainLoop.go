package main

import (
	"time"
)

const (
	ROUND_LENGTH_uS  = 250000 //0.25s
	CONNECT_THROTTLE = time.Millisecond
)

func mainLoop() {
	var tickNum uint64

	roundTime := time.Duration(ROUND_LENGTH_uS * time.Microsecond)

	for serverState == SERVER_RUNNING {
		tickNum++
		start := time.Now()

		descLock.Lock()
		var newList []*descData
		//Remove dead descriptors
		for _, desc := range descList {
			if desc.state == CON_DISCONNECTED {
				errLog("Removed %v", desc.id)
				continue
			}
			newList = append(newList, desc)
		}
		descList = newList

		for _, desc := range descList {
			desc.interp()
		}
		descLock.Unlock()

		//Sleep for remaining round time
		since := roundTime - time.Since(start)
		time.Sleep(since)
		//fmt.Println(since.String())
	}
}
