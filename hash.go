package main

import (
	"sync"
	"time"

	"github.com/remeh/sizedwaitgroup"
	"golang.org/x/crypto/bcrypt"
)

const (
	HASH_SLEEP           = time.Millisecond * 10
	HASH_TIMEOUT         = time.Second * 60
	PASSPHRASE_HASH_COST = 14
	HASH_DEPTH_MAX       = 100
)

var hashDepth int

type hashData struct {
	isTest bool
	id     uint64
	desc   *descData

	pass       []byte
	hash       []byte
	doEncrypt  bool
	changePass bool

	complete    bool
	failed      bool
	correctPass bool

	started     time.Time
	workStarted time.Time
}

var hashList []*hashData
var hashLock sync.Mutex

func hashReceiver() {

	hashLock.Lock()
	defer hashLock.Unlock()

	var newList []*hashData
	var newCount int
	for _, item := range hashList {
		if item.complete && !item.failed {
			if item.changePass {
				changePass(item)
				continue
			}
			if item.doEncrypt {
				hashGenComplete(item)
			} else {
				passCheckComplete(item)
			}
			continue

		} else if item.failed {
			hashGenFail(item)
			continue

		} else if !item.workStarted.IsZero() && time.Since(item.workStarted) > HASH_TIMEOUT {
			item.desc.send("The passphrase processing timed out. Sorry!")
			critLog("#%v: passphrase hashing timed out...", item.id)
			continue
		}

		newList = append(newList, item)
		newCount++
	}
	hashList = newList
	hashDepth = newCount
}

// Threaded async hash
func hasherDaemon() {
	wg := sizedwaitgroup.New(numThreads)

	for serverState.Load() == SERVER_RUNNING {
		time.Sleep(HASH_SLEEP)

		hashLock.Lock()
		hashDepth := len(hashList)

		//Empty, just exit
		if hashDepth == 0 {
			hashLock.Unlock()
			continue
		}

		//Limit worksize to workload and threads
		workSize := numThreads
		if workSize > hashDepth {
			workSize = hashDepth
		}

		//Copy pointer list, so we can work async
		workList := hashList
		hashLock.Unlock()

		//Process worklist
		for x := 0; x < workSize; x++ {
			item := workList[(hashDepth-1)-x]
			if item.desc != nil {
				wg.Add()
				go processHash(item, &wg)
			}
		}
		wg.Wait()
		//Wait until all threads return
	}
}

func changePass(item *hashData) {
	if item.doEncrypt {
		item.desc.account.PassHash = item.hash

		//Save account
		if !item.desc.account.saveAccount() {
			//Save failure
			item.desc.sendln(warnBuf)
			item.desc.sendln("Unable to save account!")
			critLog("#%v unable to save account!", item.id)
			item.desc.close()
			return
		}

		item.desc.sendln("Passphrase changed and saved.")
		item.desc.state = CON_CHAR_LIST
		showStatePrompt(item.desc)
	} else {
		if item.correctPass {
			item.desc.state = CON_CHANGE_PASS_NEW
		} else {
			item.desc.sendln("Incorrect passphrase, try again.")
			item.desc.state = CON_CHANGE_PASS_OLD
		}
		showStatePrompt(item.desc)
	}

}

func processHash(item *hashData, wg *sizedwaitgroup.SizedWaitGroup) {
	var err error

	defer wg.Done()

	passGood := false
	if item.doEncrypt {
		item.hash, err = bcrypt.GenerateFromPassword([]byte(item.pass), PASSPHRASE_HASH_COST)
	} else {
		if bcrypt.CompareHashAndPassword(item.hash, []byte(item.pass)) == nil {
			passGood = true
		}
	}

	if passGood {
		item.correctPass = true
	}

	if err != nil {
		item.failed = true
		critLog("ERROR: #%v passphrase hashing failed!!!: %v", item.id, err.Error())
		return
	}
	item.complete = true
}

func passCheckComplete(item *hashData) {
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
		pCharList(item.desc)
	} else {
		item.desc.send("Incorrect passphrase.")
		critLog("#%v tried an incorrect passphrase.", item.id)
		item.desc.state = CON_DISCONNECTED
		item.desc.valid = false
	}
}

func hashGenComplete(item *hashData) {
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
		item.desc.sendln(warnBuf)
		item.desc.sendln("Unable to create account!")
		critLog("#%v unable to create account!", item.id)
		item.desc.close()
		return
	}

	//Save account
	if !item.desc.account.saveAccount() {
		//Save failure
		item.desc.send(warnBuf)
		item.desc.sendln("Unable to save account!")
		critLog("#%v unable to save account!", item.id)
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
	showStatePrompt(item.desc)
}

func hashGenFail(item *hashData) {
	item.desc.send("Somthing went wrong processing your passphrase. Sorry!")
	critLog("#%v passphrase hash failed!", item.id)
}
