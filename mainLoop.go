package main

import (
	"time"

	"golang.org/x/exp/rand"
)

const (
	ROUND_LENGTH_uS  = 250000                //0.25s
	CONNECT_THROTTLE = time.Millisecond * 10 //100 connections per second
	HASH_SLEEP       = time.Millisecond * 100
	HASH_TIMEOUT     = time.Second * 30
)

func mainLoop() {
	var tickNum uint64

	roundTime := time.Duration(ROUND_LENGTH_uS * time.Microsecond)

	for serverState.Load() == SERVER_RUNNING {
		tickNum++
		start := time.Now()

		descLock.Lock()
		hashReceiver()

		//Remove dead characters
		var newCharacterList []*characterData
		for _, target := range characterList {
			if !target.valid {
				target.sendToPlaying("%v slowly fades away.", target.Name)
				errLog("Removed character %v", target.Name)
				continue
			} else if time.Since(target.idleTime) > CHARACTER_IDLE {
				target.send("Idle too long, quitting...")
				target.quit(true)
				continue
			}
			newCharacterList = append(newCharacterList, target)
		}
		characterList = newCharacterList

		//Remove dead descriptors
		var newDescList []*descData
		for _, desc := range descList {
			if desc.state == CON_LOGIN &&
				time.Since(desc.idleTime) > LOGIN_AFK {
				desc.sendln("\r\nIdle too long, disconnecting.")
				desc.killDesc()
				continue
			} else if desc.state != CON_PLAYING &&
				time.Since(desc.idleTime) > AFK_DESC {
				desc.sendln("\r\nIdle too long, disconnecting.")
				desc.killDesc()
				continue
			} else if desc.state == CON_DISCONNECTED || !desc.valid {
				desc.killDesc()
				continue
			}
			newDescList = append(newDescList, desc)
		}
		descList = newDescList

		//Interpret all
		for _, desc := range descList {
			desc.interp()
		}

		//Shuffle descriptor list
		numDesc := len(descList) - 1
		if numDesc > 1 {
			j := rand.Intn(numDesc) + 1
			descList[0], descList[j] = descList[j], descList[0]
		}
		descLock.Unlock()

		//Sleep for remaining round time
		timeLeft := roundTime - time.Since(start)

		if timeLeft <= 0 {
			errLog("Round went over: %v", time.Duration(timeLeft).Round(time.Microsecond).Abs().String())
		} else {
			time.Sleep(timeLeft)
		}
	}
}

func (desc *descData) killDesc() {
	errLog("Removed #%v", desc.id)
	desc.valid = false
	desc.conn.Close()
	if desc.character != nil {
		desc.character.desc = nil
	}
}
