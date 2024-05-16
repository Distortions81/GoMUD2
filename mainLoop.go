package main

import (
	"time"

	"github.com/remeh/sizedwaitgroup"
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
			critLog("Round went over: %v", time.Duration(timeLeft).Truncate(time.Microsecond).Abs().String())
		} else {
			time.Sleep(timeLeft)
		}
	}
}

func sendOutput() {
	//multi-thread output processing
	wg := sizedwaitgroup.New(numThreads)

	for _, desc := range descList {
		if desc.haveOut {
			wg.Add()
			go func(tdesc *descData) {

				//Character map translation
				if !tdesc.telnet.options.UTF {
					tdesc.outBuf = encodeFromUTF(tdesc.telnet.charMap, tdesc.outBuf)
				}

				//Color
				tdesc.outBuf = ANSIColor(tdesc.outBuf)

				//Add telnet go-ahead if enabled, and there is no newline
				if tdesc.telnet.options != nil && !tdesc.telnet.options.SUPGA {
					if tdesc.outBuf[len(tdesc.outBuf)-1] != '\n' {
						tdesc.outBuf = append(tdesc.outBuf, []byte{TermCmd_IAC, TermCmd_GOAHEAD}...)
					}
				}

				_, err := tdesc.conn.Write(tdesc.outBuf)
				if err != nil {
					errLog("#%v: %v: write failed (connection lost)", tdesc.id, tdesc.cAddr)
					tdesc.state = CON_DISCONNECTED
					tdesc.valid = false
				}

				tdesc.outBuf = []byte{}
				tdesc.haveOut = false
				wg.Done()
			}(desc)
		}
	}
	wg.Wait()
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
			if target.desc != nil {
				target.saveCharacter()
			}
			target.leaveRoom()
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

		} else if desc.state <= CON_CHECK_PASS &&
			time.Since(desc.idleTime) > LOGIN_IDLE {
			desc.sendln("\r\nIdle too long, disconnecting.")
			desc.killDesc()
			continue

		} else if desc.state == CON_DISCONNECTED ||
			!desc.valid {
			desc.killDesc()
			continue

		} else if desc.state != CON_PLAYING &&
			time.Since(desc.idleTime) > MENU_IDLE {
			if desc.character != nil && desc.character.Level < 0 {
				continue
			}
			desc.sendln("\r\nIdle too long, disconnecting.")
			desc.killDesc()
			continue
		} else {
			newDescList = append(newDescList, desc)
		}
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
