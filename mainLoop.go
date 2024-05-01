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
		//Remove dead players
		var newPlayList []*playerData
		for _, target := range playList {
			if !target.valid {
				errLog("Removed player %v", target.Name)
				continue
			}
			newPlayList = append(newPlayList, target)
		}
		playList = newPlayList

		//Remove dead descriptors
		var newDescList []*descData
		for _, desc := range descList {
			if desc.state == CON_DISCONNECTED || !desc.valid {
				errLog("Removed #%v", desc.id)
				continue
			}
			newDescList = append(newDescList, desc)
		}
		descList = newDescList

		//Interpret all
		for _, desc := range descList {
			desc.interp()
		}
		descLock.Unlock()

		//Sleep for remaining round time
		since := roundTime - time.Since(start)
		time.Sleep(since)

		took := time.Duration(since).Round(time.Millisecond)
		if took < (roundTime / 10) {
			errLog("Over 90%% load! Round took %v", took.String())
		}
	}
}
