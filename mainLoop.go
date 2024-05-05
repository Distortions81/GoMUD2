package main

import (
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"
)

const (
	ROUND_LENGTH_uS  = 250000                //0.25s
	CONNECT_THROTTLE = time.Millisecond * 10 //100 connections per second
	HASH_SLEEP       = time.Millisecond * 100
	HASH_TIMEOUT     = time.Second * 30
)

type toHashData struct {
	id   uint64
	desc *descData

	pass []byte
	hash []byte

	complete bool
	failed   bool
	started  time.Time
}

var hashList []*toHashData
var hashLock sync.Mutex

func hashLoop() {
	for serverState.Load() == SERVER_RUNNING {
		time.Sleep(HASH_SLEEP)
		hashLock.Lock()

		hashListLen := len(hashList)
		if hashListLen == 0 {
			hashLock.Unlock()
			continue
		}

		item := hashList[0]
		if item.complete || item.failed {
			hashLock.Unlock()
			continue
		}

		var err error
		start := time.Now()
		hashLock.Unlock()
		item.hash, err = bcrypt.GenerateFromPassword([]byte(item.pass), PASSPHRASE_HASH_COST)
		hashLock.Lock()
		errLog("Password hash took %v.", time.Since(start).Round(time.Microsecond))
		if err != nil {
			item.failed = true
			critLog("ERROR: #%v password hashing failed!!!: %v", item.id, err.Error())
			hashLock.Unlock()
			continue
		}
		item.complete = true

		hashLock.Unlock()
	}
}

func mainLoop() {
	var tickNum uint64

	roundTime := time.Duration(ROUND_LENGTH_uS * time.Microsecond)

	for serverState.Load() == SERVER_RUNNING {
		tickNum++
		start := time.Now()

		descLock.Lock()
		hashLock.Lock()

		hashLen := len(hashList)

		if hashLen > 0 {
			item := hashList[0]

			if item.complete && !item.failed {
				item.desc.sendln("Hashing complete!")
				if hashLen > 1 {
					hashList = hashList[1:]
				} else {
					hashList = []*toHashData{}
				}

				err := item.desc.account.createAccountDir()
				if err != nil {
					item.desc.send(warnBuf)
					item.desc.sendln("Unable to create account! Please let moderators know!")
					errLog("#%v unable to create account!", item.id)
					item.desc.close()
					hashLock.Lock()
					continue
				}

				notSaved := item.desc.account.saveAccount()
				if notSaved {
					item.desc.send(warnBuf)
					item.desc.sendln("Unable to save account! Please let moderators know!")
					errLog("#%v unable to save account!", item.id)
					item.desc.close()
					hashLock.Lock()
					continue
				} else {
					item.desc.sendln("Account created and saved.")
					newAcc := &accountIndexData{
						Login:       item.desc.account.Login,
						Fingerprint: item.desc.account.Fingerprint,
						Added:       time.Now(),
					}
					accountIndex[item.desc.account.Login] = newAcc
					saveAccountIndex()
				}

				item.desc.state = CON_CHAR_LIST
				gCharList(item.desc)

			} else if item.failed {
				item.desc.send("Somthing went wrong hashing your password. Sorry!")
				errLog("#%v password hash returned not complete or failed!", item.id)
				if hashLen > 1 {
					hashList = hashList[1:]
				} else {
					hashList = []*toHashData{}
				}
			}
			if time.Since(item.started) > HASH_TIMEOUT {
				item.desc.send("The password hashing timed out. Sorry!")
				errLog("#%v: Password hashing timed out...", item.id)
				if hashLen > 1 {
					hashList = hashList[1:]
				} else {
					hashList = []*toHashData{}
				}
			}
		}
		hashLock.Unlock()

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
