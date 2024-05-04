package main

import (
	"time"
)

const (
	ROUND_LENGTH_uS  = 250000                //0.25s
	CONNECT_THROTTLE = time.Millisecond * 10 //100 connections per second
)

func mainLoop() {
	var tickNum uint64

	roundTime := time.Duration(ROUND_LENGTH_uS * time.Microsecond)

	for serverState.Load() == SERVER_RUNNING {
		tickNum++
		start := time.Now()

		descLock.Lock()
		//Remove dead characters
		var newCharacterList []*characterData
		for _, target := range characterList {
			if !target.valid {
				target.sendToPlaying("%v slowly fades away.", target.Name)
				errLog("Removed character %v", target.Name)
				continue
			}
			newCharacterList = append(newCharacterList, target)
		}
		characterList = newCharacterList

		//Remove dead descriptors
		var newDescList []*descData
		for _, desc := range descList {
			if desc.state == CON_DISCONNECTED || !desc.valid {
				errLog("Removed #%v", desc.id)
				desc.conn.Close()
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
		timeLeft := roundTime - time.Since(start)

		if timeLeft <= 0 {
			errLog("Round took %v", time.Duration(timeLeft).String())
		} else {
			time.Sleep(timeLeft)
		}
	}
}
