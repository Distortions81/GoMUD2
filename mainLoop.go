package main

import (
	"time"
)

const (
	ROUND_LENGTH_uS  = 250000 //0.25s
	CONNECT_THROTTLE = time.Millisecond
	LAG_THRESH       = time.Millisecond
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

		since := time.Since(start)
		if since > time.Millisecond {
			errLog("Round took %v", since.Round(LAG_THRESH).String())
		}

		//Sleep for remaining round time
		timeLeft := roundTime - time.Since(start)
		time.Sleep(timeLeft)

	}
}
