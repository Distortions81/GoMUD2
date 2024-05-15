package main

import (
	"time"

	"golang.org/x/exp/rand"
)

const (
	ROUND_LENGTH_uS  = 250000 //0.25s
	CONNECT_THROTTLE = time.Microsecond * 200
)

func mainLoop() {
	var tickNum uint64

	roundTime := time.Duration(ROUND_LENGTH_uS * time.Microsecond)

	for serverState.Load() == SERVER_RUNNING {
		tickNum++
		start := time.Now()

		descLock.Lock()
		removeDeadDesc()
		removeDeadChar()
		hashReceiver()
		descShuffle()
		interpAllDesc()
		saveAllAreas(true)
		sendOutput()
		descLock.Unlock()

		//Sleep for remaining round time
		timeLeft := roundTime - time.Since(start)
		if timeLeft <= 0 {
			critLog("Round went over: %v", time.Duration(timeLeft).Round(time.Microsecond).Abs().String())
		} else {
			time.Sleep(timeLeft)
		}
	}
}

func sendOutput() {
	for _, desc := range descList {
		if desc.haveOut {

			//Character map translation
			if !desc.telnet.options.UTF {
				desc.outBuf = encodeFromUTF(desc.telnet.charMap, desc.outBuf)
			}

			//Color
			desc.outBuf = ANSIColor(desc.outBuf)

			//Add telnet go-ahead if enabled, and there is no newline
			if desc.telnet.options != nil && !desc.telnet.options.SUPGA {
				if desc.outBuf[len(desc.outBuf)-1] != '\n' {
					desc.outBuf = append(desc.outBuf, []byte{TermCmd_IAC, TermCmd_GOAHEAD}...)
				}
			}

			_, err := desc.conn.Write(desc.outBuf)
			if err != nil {
				//errLog("#%v: %v: write failed (connection lost)", desc.id, desc.cAddr)
				desc.state = CON_DISCONNECTED
				desc.valid = false
			}

			desc.outBuf = []byte{}
			desc.haveOut = false
		}
	}
}

func descShuffle() {
	//Shuffle descriptor list
	numDesc := len(descList) - 1
	if numDesc > 1 {
		j := rand.Intn(numDesc) + 1
		descList[0], descList[j] = descList[j], descList[0]
	}
}

func interpAllDesc() {
	//Interpret all
	for _, desc := range descList {
		desc.interp()
	}
}

func removeDeadChar() {
	//Remove dead characters
	var newCharacterList []*characterData
	for _, target := range charList {
		if !target.valid {
			target.sendToPlaying("%v slowly fades away.", target.Name)
			errLog("Removed character %v from charList.", target.Name)
			target.saveCharacter()
			target.fromRoom()
			continue
		} else if time.Since(target.idleTime) > CHARACTER_IDLE {
			target.send("Idle too long, quitting...")
			target.quit(true)
			continue
		}
		newCharacterList = append(newCharacterList, target)
	}
	charList = newCharacterList
}

func removeDeadDesc() {
	//Remove dead descriptors
	var newDescList []*descData
	for _, desc := range descList {
		if desc.state == CON_HASH_WAIT {
			newDescList = append(newDescList, desc)
			continue

		} else if desc.state == CON_LOGIN &&
			time.Since(desc.idleTime) > LOGIN_AFK {
			desc.sendln("\r\nIdle too long, disconnecting.")
			desc.killDesc()
			continue

		} else if desc.state == CON_DISCONNECTED ||
			!desc.valid {
			desc.killDesc()
			continue

		} else if desc.state != CON_PLAYING &&
			time.Since(desc.idleTime) > AFK_DESC {
			if desc.character != nil && desc.character.Level < 0 {
				continue
			}
			desc.sendln("\r\nIdle too long, disconnecting.")
			desc.killDesc()
			continue
		}
		newDescList = append(newDescList, desc)
	}
	descList = newDescList
}

func (desc *descData) killDesc() {
	//errLog("Removed #%v", desc.id)
	desc.valid = false
	desc.conn.Close()
	if desc.character != nil {
		desc.character.desc = nil
	}
}
