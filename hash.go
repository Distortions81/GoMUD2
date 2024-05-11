package main

import (
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	HASH_SLEEP           = time.Millisecond * 10
	HASH_TIMEOUT         = time.Second * 60
	PASSPHRASE_HASH_COST = 12
	HASH_DEPTH_MAX       = 100
)

var lastHashTime time.Duration = time.Second * 5
var lastHashCheckTime time.Duration = time.Second * 5

type toHashData struct {
	isTest bool
	id     uint64
	desc   *descData

	pass      []byte
	hash      []byte
	doEncrypt bool

	complete    bool
	failed      bool
	correctPass bool

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
			if item.doEncrypt {
				hashGenComplete(item)
			} else {
				passCheckComplete(item)
			}
			removeFirstHash()

		} else if item.failed {
			hashGenFail(item)
			removeFirstHash()

		} else if !item.workStarted.IsZero() && time.Since(item.workStarted) > HASH_TIMEOUT {
			item.desc.send("The passphrase processing timed out. Sorry!")
			errLog("#%v: passphrase hashing timed out...", item.id)
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

		passGood := false
		if item.doEncrypt {
			item.hash, err = bcrypt.GenerateFromPassword([]byte(item.pass), PASSPHRASE_HASH_COST)
		} else {
			if bcrypt.CompareHashAndPassword(item.hash, []byte(item.pass)) == nil {
				passGood = true
			} else {
				item.desc.sendln("Incorrect passphrase.")
				critLog("#%v: tried a invalid passphrase!", item.id)
			}
		}

		hashLock.Lock()

		if item.doEncrypt {
			took := time.Since(start).Round(time.Millisecond)
			errLog("passphrase hash took %v.", took)
			lastHashTime = took
		} else {
			took := time.Since(start).Round(time.Millisecond)
			errLog("passphrase check took %v.", took)
			if passGood {
				item.correctPass = true
			}
			lastHashCheckTime = took
		}

		if err != nil {
			item.failed = true
			critLog("ERROR: #%v passphrase hashing failed!!!: %v", item.id, err.Error())
			hashLock.Unlock()
			continue
		}
		item.complete = true

		hashLock.Unlock()
	}
}

func passCheckComplete(item *toHashData) {
	if item.isTest {
		return
	}
	if item.desc == nil {
		critLog("Player left before passphrase check finished.")
		return
	}
	if item.correctPass {
		//Send to char menu
		item.desc.state = CON_CHAR_LIST
		gCharList(item.desc)
	} else {
		item.desc.send("Invalid passphrase.")
		item.desc.state = CON_DISCONNECTED
		item.desc.valid = false
	}
}

func hashGenComplete(item *toHashData) {
	if item.isTest {
		return
	}
	if item.desc == nil {
		critLog("Player left before passphrase hash finished.")
		return
	}
	item.desc.sendln("Passphrase processing complete!")
	item.desc.account.PassHash = item.hash

	//Check again! We don't want a collision
	//Otherwise could be exploitable.
	if !isAccNameAvail(item.desc.account.Login) {
		item.desc.send("Sorry, that login name is already in use.")
		item.desc.valid = false
		item.desc.state = CON_DISCONNECTED
		return
	}

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
	if !item.desc.account.saveAccount() {
		//Save failure
		item.desc.send(warnBuf)
		item.desc.sendln("Unable to save account!")
		errLog("#%v unable to save account!", item.id)
		item.desc.close()
		return
	}

	newAcc := &accountIndexData{
		Login: item.desc.account.Login,
		UUID:  item.desc.account.UUID,
		Added: time.Now().UTC(),
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
	item.desc.send("Somthing went wrong processing your passphrase. Sorry!")
	errLog("#%v passphrase hash failed!", item.id)
}

func removeFirstHash() {
	if len(hashList) > 1 {
		hashList = hashList[1:]
	} else {
		hashList = []*toHashData{}
	}
}
