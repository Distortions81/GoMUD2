package main

import (
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type toHashData struct {
	id   uint64
	desc *descData

	pass      []byte
	hash      []byte
	doEncrypt bool

	complete bool
	failed   bool
	started  time.Time
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

		} else if time.Since(item.started) > HASH_TIMEOUT {
			item.desc.send("The password hashing timed out. Sorry!")
			errLog("#%v: Password hashing timed out...", item.id)
			removeFirstHash()
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

func hashGenComplete(item *toHashData) {
	item.desc.sendln("Hashing complete!")

	//Create account
	err := item.desc.account.createAccountDir()
	if err != nil {
		item.desc.send(warnBuf)
		item.desc.sendln("Unable to create account! Please let moderators know!")
		errLog("#%v unable to create account!", item.id)
		item.desc.close()
		hashLock.Lock()
		return
	}

	//Save account
	notSaved := item.desc.account.saveAccount()
	if notSaved {

		//Save failure
		item.desc.send(warnBuf)
		item.desc.sendln("Unable to save account! Please let moderators know!")
		errLog("#%v unable to save account!", item.id)
		item.desc.close()
		hashLock.Lock()
		return
	} else {

		//save success
		item.desc.sendln("Account created and saved.")
		newAcc := &accountIndexData{
			Login:       item.desc.account.Login,
			Fingerprint: item.desc.account.Fingerprint,
			Added:       time.Now(),
		}

		//Update acc index
		accountIndex[item.desc.account.Login] = newAcc
		saveAccountIndex()
		item.desc.account.saveAccount()
	}

	//Send to char menu
	item.desc.state = CON_CHAR_LIST
	gCharList(item.desc)
}

func hashGenFail(item *toHashData) {
	item.desc.send("Somthing went wrong hashing your password. Sorry!")
	errLog("#%v password hash returned not complete or failed!", item.id)
}

func removeFirstHash() {
	if len(hashList) > 1 {
		hashList = hashList[1:]
	} else {
		hashList = []*toHashData{}
	}
}
