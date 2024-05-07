package main

import (
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	HASH_SLEEP           = time.Millisecond * 100
	HASH_TIMEOUT         = time.Second * 60
	PASSPHRASE_HASH_COST = 15
	HASH_DEPTH_MAX       = 100
)

var lastHashTime time.Duration = time.Second * 5

type toHashData struct {
	isTest bool
	id     uint64
	desc   *descData

	pass      []byte
	hash      []byte
	doEncrypt bool

	complete bool
	failed   bool

	started     time.Time
	workStarted time.Time
}

var hashList []*toHashData
var hashLock sync.Mutex

func hashReceiver() {
	hashLock.Lock()

	hashLen := len(hashList)

	if hashLen > 0 {
		item := hashList[0]

		if item.complete && !item.failed {
			hashGenComplete(item)
			removeFirstHash()

		} else if item.failed {
			hashGenFail(item)
			removeFirstHash()

		} else if !item.workStarted.IsZero() && time.Since(item.workStarted) > HASH_TIMEOUT {
			item.desc.send("The password processing timed out. Sorry!")
			errLog("#%v: Password hashing timed out...", item.id)
			removeFirstHash()
			item.desc.close()
		}

	}
	hashLock.Unlock()
}

func hasherDaemon() {
	for serverState.Load() == SERVER_RUNNING {
		time.Sleep(HASH_SLEEP)
		hashLock.Lock()

		if len(hashList) == 0 {
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
		item.workStarted = time.Now()
		hashLock.Unlock()

		item.hash, err = bcrypt.GenerateFromPassword([]byte(item.pass), PASSPHRASE_HASH_COST)

		hashLock.Lock()

		took := time.Since(start).Round(time.Millisecond)
		errLog("Password hash took %v.", took)
		lastHashTime = took

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

func hashGenComplete(item *toHashData) {
	if item.isTest {
		return
	}
	if item.desc == nil {
		critLog("Player left before password hash finished.")
		return
	}
	item.desc.sendln("Passphrase processing complete!")
	item.desc.account.PassHash = item.hash

	//Create account
	err := item.desc.account.createAccountDir()
	if err != nil {
		item.desc.send(warnBuf)
		item.desc.sendln("Unable to create account!")
		errLog("#%v unable to create account!", item.id)
		item.desc.close()
		return
	}

	//Save account
	notSaved := item.desc.account.saveAccount()
	if notSaved {

		//Save failure
		item.desc.send(warnBuf)
		item.desc.sendln("Unable to save account!")
		errLog("#%v unable to save account!", item.id)
		item.desc.close()
		return
	}

	newAcc := &accountIndexData{
		Login:       item.desc.account.Login,
		Fingerprint: item.desc.account.Fingerprint,
		Added:       time.Now().UTC(),
	}

	//Update acc index
	accountIndex[item.desc.account.Login] = newAcc
	saveAccountIndex()

	//save success
	item.desc.sendln("Account created and saved.")

	//Send to char menu
	item.desc.state = CON_CHAR_LIST
	gCharList(item.desc)
}

func hashGenFail(item *toHashData) {
	item.desc.send("Somthing went wrong processing your password. Sorry!")
	errLog("#%v password hash failed!", item.id)
}

func removeFirstHash() {
	if len(hashList) > 1 {
		hashList = hashList[1:]
	} else {
		hashList = []*toHashData{}
	}
}
