package main

import (
	"fmt"
	"math"
	"time"

	"github.com/remeh/sizedwaitgroup"
	"golang.org/x/exp/rand"
)

const (
	PULSE_PER_SECOND = 3.0
	PULSE_PER_MINUTE = PULSE_PER_SECOND * 60
	PULSE_LENGTH_uS  = 1000000 / PULSE_PER_SECOND

	INTERP_LOOP_MARGIN  = (INTERP_LOOP_REST_uS * 2) * time.Microsecond
	INTERP_LOOP_REST_uS = 4000

	PULSE_HISTORY_LEN  = 5 * PULSE_PER_MINUTE
	FORCESAVE_INTERVAL = 5 * PULSE_PER_MINUTE
)

var loopTask int
var fullPulseHistory, partialPulseHistory []int64
var peakFullPulse, peakPartialPulse int64

var historyLen int

func mainLoop() {
	var tickNum uint64

	pulseTime := time.Duration(int(math.Round(PULSE_LENGTH_uS))) * time.Microsecond
	loopTime := time.Duration(INTERP_LOOP_REST_uS * time.Microsecond)

	//errLog("Pulse time: %v", pulseTime.Abs().String())

	for serverState == SERVER_RUNNING {
		tickNum++
		start := time.Now().UTC()

		descLock.Lock()

		hashReceiver()

		//Force save all
		if tickNum%FORCESAVE_INTERVAL == 0 {
			saveAllCharacters(true)
		} else {
			switch loopTask {
			case 0:
				expireBlocks()
			case 1:
				saveNotes(false)
			case 2:
				writeBlocked(false)
			case 3:
				saveAllAreas(false)
			case 4:
				saveAllCharacters(false)
			case 5:
				removeDeadDesc()
			case 6:
				removeDeadChar()
			case 7:
				descShuffle()
			case 8:
				updateRecentLogins()
			case 9:
				writeBugs()
				loopTask = 0
			}
			loopTask++
		}

		resetProcessed()
		interpAllDesc()
		sendOutput()

		/* Record remaining time */
		ppulse := time.Since(start).Microseconds()
		partialPulseHistory = append(partialPulseHistory, ppulse)
		if peakPartialPulse < ppulse {
			peakPartialPulse = ppulse
		}

		/* Instant command response */
		/* Burns all free frame time looking for incoming commands */
		if *instantRespond && len(descList) > 0 {
			for {
				loopStart := time.Now().UTC()
				for _, desc := range descList {
					if !desc.valid {
						continue
					}
					if desc.processed {
						continue
					}
					if desc.interp() {
						desc.processed = true

						/*
							timeLeft := pulseTime - time.Since(start)
							if timeLeft < INTERP_LOOP_MARGIN {
								break
							}
						*/
					}

				}
				sendOutput()

				timeLeft := pulseTime - time.Since(start)
				if timeLeft < INTERP_LOOP_MARGIN {
					break
				}

				//Sleep for remaining interp loop time
				loopLeft := loopTime - time.Since(loopStart)
				time.Sleep(loopLeft)
			}
		}
		descLock.Unlock()

		//Sleep for remaining pulse time
		took := time.Since(start)
		timeLeft := pulseTime - took

		/* Record remaining time */
		tookMicro := took.Microseconds()
		fullPulseHistory = append(fullPulseHistory, tookMicro)
		if peakFullPulse < tookMicro {
			peakFullPulse = tookMicro
		}

		/* Trim to max history */
		if historyLen >= PULSE_HISTORY_LEN-1 {
			fullPulseHistory = fullPulseHistory[:PULSE_HISTORY_LEN-1]
			partialPulseHistory = partialPulseHistory[:PULSE_HISTORY_LEN-1]
		} else {
			historyLen++
		}

		//Alert if we went more than 1% over frame time
		pul := int(math.Round(float64(PULSE_LENGTH_uS) / 100.0))
		if timeLeft < -time.Duration(pul)*time.Microsecond {
			critLog("Pulse lag: %v", time.Duration(timeLeft).Truncate(time.Microsecond).Abs().String())
		} else {
			//critLog("Pulse left: %v", time.Duration(timeLeft).Truncate(time.Microsecond).Abs().String())
			time.Sleep(timeLeft)
		}
	}
}

func resetProcessed() {
	for _, desc := range descList {
		if desc.processed {
			desc.processed = false
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
				tdesc.doOutput()
				wg.Done()
			}(desc)
		}
	}
	wg.Wait()
}

const timeStampFormat = "(3:4:05pm MST):"

func (tdesc *descData) doOutput() {

	//Timestamp
	if tdesc.character != nil && tdesc.character.Config.hasFlag(CONFIG_TIMESTAMP) {
		ts := fmt.Sprintf("%v", time.Now().UTC().Format(timeStampFormat)) + NEWLINE
		tdesc.outBuf = append([]byte(ts), tdesc.outBuf...)
	}

	//Emoji
	if tdesc.character != nil && tdesc.character.Config.hasFlag(CONFIG_TEXT_EMOJI) {
		tdesc.outBuf = unicodeToName(tdesc.outBuf)
	} else {
		if tdesc.telnet.Options != nil && tdesc.telnet.Options.MTTS.hasFlag(MTTS_UTF8) {
			tdesc.outBuf = nameToUnicode(tdesc.outBuf)
		} else {
			tdesc.outBuf = unicodeToName(tdesc.outBuf)
		}
	}

	//Color
	if tdesc.telnet.Options != nil {
		if !tdesc.telnet.Options.MTTS.hasFlag(MTTS_ANSI) ||
			(tdesc.character != nil && tdesc.character.Config.hasFlag(CONFIG_NOCOLOR)) {
			tdesc.outBuf = ColorRemove(tdesc.outBuf)
		} else if tdesc.telnet.Options.MTTS.hasFlag(MTTS_256) {
			tdesc.outBuf = ANSIColor(tdesc.outBuf, COLOR_256)
		} else if tdesc.telnet.Options.MTTS.hasFlag(MTTS_TRUECOLOR) {
			tdesc.outBuf = ANSIColor(tdesc.outBuf, COLOR_TRUE)
		} else {
			tdesc.outBuf = ANSIColor(tdesc.outBuf, COLOR_16)
		}
	} else {
		tdesc.outBuf = ANSIColor(tdesc.outBuf, COLOR_16)
	}

	//Character map translation
	if !tdesc.telnet.Options.MTTS.hasFlag(MTTS_UTF8) {
		tdesc.outBuf = encodeFromUTF(tdesc.telnet.charMap, tdesc.outBuf)
	}

	//Add telnet go-ahead if enabled, and there is no newline ending
	if tdesc.telnet.Options != nil && !tdesc.telnet.Options.SuppressGoAhead {
		outLen := len(tdesc.outBuf) - 1
		if outLen > 0 {
			if tdesc.outBuf[outLen-1] != '\n' {
				tdesc.outBuf = append(tdesc.outBuf, []byte{TermCmd_IAC, TermCmd_GOAHEAD}...)
			}
		}
	}

	if tdesc.state == CON_PLAYING {
		if target := tdesc.character; target != nil {
			if target.OLCEditor.OLCMode != OLC_NONE {
				flag := ""
				if target.Config.hasFlag(1 << CONFIG_OLCHERE) {
					flag = " (OLCHere On)"
				}

				var avnum, rvnum int
				if target.OLCEditor.area != nil {
					avnum = target.OLCEditor.area.VNUM
				}
				if target.OLCEditor.room != nil {
					rvnum = target.OLCEditor.room.VNUM
				}

				buf := fmt.Sprintf("<OLC %v: (%v:%v) help, exit%v>:"+NEWLINE,
					olcModes[target.OLCEditor.OLCMode].name,
					avnum, rvnum,
					flag)
				tdesc.outBuf = append(tdesc.outBuf, []byte(buf)...)
			} else if target.DraftNotes.Editing {
				buf := ""
				if target.Level >= LEVEL_MODERATOR {
					buf = fmt.Sprintf(noteDraftPromptMod + NEWLINE)
				} else {
					buf = fmt.Sprintf(noteDraftPrompt + NEWLINE)
				}
				tdesc.outBuf = append(tdesc.outBuf, []byte(buf)...)
			} else if target.Prompt != "" {
				buf := fmt.Sprintf("<%v>: "+NEWLINE, target.Prompt)
				tdesc.outBuf = append(tdesc.outBuf, []byte(buf)...)
			}
		}
	}

	_, err := tdesc.conn.Write(tdesc.outBuf)
	if err != nil {
		mudLog("#%v: %v: write failed (connection lost)", tdesc.id, tdesc.ip)
		if tdesc.character != nil && tdesc.character.valid {
			tdesc.character.sendToRoom("%v has lost their connection.", tdesc.character.Name)
		}
	}

	tdesc.outBuf = []byte{}
	tdesc.haveOut = false
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
		if !desc.valid {
			continue
		}
		if desc.processed {
			continue
		}
		if desc.interp() {
			desc.processed = true
		}
	}
}

func removeDeadChar() {
	//Remove dead characters
	var newCharacterList []*characterData
	for _, target := range charList {
		if target.desc == nil || (target.desc != nil && !target.desc.valid) &&
			time.Since(target.idleTime) > NO_LINK_TIME {
			target.quit(true)
			continue
		} else if !target.valid {
			errLog("Removed character %v from charList.", target.Name)
			continue
		} else if target.Level >= LEVEL_BUILDER &&
			time.Since(target.idleTime) > BUILDER_IDLE {
			target.send("Idle too long, quitting...")
			target.quit(true)
			continue

		} else if target.Level < LEVEL_BUILDER &&
			time.Since(target.idleTime) > CHARACTER_IDLE {
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
		if desc.state == CON_DISCONNECTED ||
			!desc.valid {
			mudLog("Removed #%v", desc.id)
			desc.killConn()
			continue
		} else if desc.state == CON_HASH_WAIT {
			//Don't do anything
		} else if desc.state <= CON_CHECK_PASS &&
			time.Since(desc.idleTime) > LOGIN_IDLE {
			desc.sendln(NEWLINE + "Idle too long, disconnecting.")
			desc.kill()
		} else if desc.state != CON_PLAYING &&
			time.Since(desc.idleTime) > MENU_IDLE {
			desc.sendln(NEWLINE + "Idle too long, disconnecting.")
			desc.kill()
		}

		newDescList = append(newDescList, desc)

	}
	descList = newDescList
}

func (desc *descData) kill() {
	if desc == nil {
		return
	}
	desc.valid = false
	desc.state = CON_DISCONNECTED
}

func (desc *descData) killConn() {
	if desc == nil {
		return
	}

	desc.valid = false
	desc.state = CON_DISCONNECTED
	desc.conn.Close()
}
