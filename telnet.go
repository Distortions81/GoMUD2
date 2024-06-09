package main

import (
	"fmt"
	"net"
	"strings"
)

const (
	MIN_TERM_WIDTH = 80
	MAX_TERM_WIDTH = 254

	MIN_TERM_HEIGHT = 20
	MAX_TERM_HEIGHT = 254
)

func sendTelnetCmds(conn net.Conn) {
	sendCmd(conn, TermCmd_DO, TermOpt_SUP_GOAHEAD)
	sendCmd(conn, TermCmd_DO, TermOpt_TERMINAL_TYPE)
	sendCmd(conn, TermCmd_WILL, TermOpt_CHARSET)
	sendCmd(conn, TermCmd_WILL, TermOpt_SUP_GOAHEAD)
	sendCmd(conn, TermCmd_DO, TermOpt_WINDOW_SIZE)
}
func (player *characterData) sendTestString() {
	player.send("Falsches Üben von Xylophonmusik quält jeden größeren Zwerg")
}

func cmdTelnet(player *characterData, input string) {
	if player.desc == nil {
		return
	}

	telnet := player.desc.telnet
	if input == "" {
		buf := "Telnet options:" + NEWLINE
		termType := "Not detected."
		if telnet.termType != "" {
			termType = telnet.termType
		}
		buf = buf + fmt.Sprintf("Selected character map: %v"+NEWLINE, telnet.Charset)
		buf = buf + fmt.Sprintf("Detected client: %v"+NEWLINE, termType)

		if telnet.Options == nil {
			telnet.Options = &termSettings{}
		}
		if telnet.Options.SuppressGoAhead {
			buf = buf + "Supressing Go-Ahead Signal (SUPGA)" + NEWLINE
		}
		if telnet.Options.NoColor {
			buf = buf + "ANSI Color disabled." + NEWLINE
		} else {
			buf = buf + "ANSI {RC{Go{Yl{Bo{Mr{x enabled." + NEWLINE
		}
		if telnet.Options.ANSI256 {
			buf = buf + "Supports [0882[094[0945[1006 [064c[028o[029l[030o[024r [018m[054o[090d[089e[088.[x" + NEWLINE
		}
		if telnet.Options.ANSI24 {
			buf = buf + "Supports 24-bit true-color" + NEWLINE
		}
		if telnet.Options.NAWS {
			buf = buf + fmt.Sprintf("Window size: %v x %v", telnet.Options.TermWidth, telnet.Options.TermHeight)
		}

		player.send(buf)
		if telnet.termType == "" {
			player.send("Options: telnet SAVE, UTF, SUPGA, COLOR or the name of a charmap.")
			player.send("To see a list of available character maps, type 'telnet charmaps'")
			player.send("telnet SAVE will save these settings to your account.")
		}
		return
	}
	if strings.EqualFold("COLOR", input) {
		if player.desc == nil {
			return
		}
		if telnet.Options.NoColor {
			player.desc.telnet.Options.NoColor = false
			player.send("ANSI color is now enabled.")
		} else {
			player.desc.telnet.Options.NoColor = true
			player.send("ANSI color is now disabled.")
		}
		return
	}
	if strings.EqualFold("256", input) {
		if player.desc == nil {
			return
		}
		if telnet.Options.ANSI256 {
			player.desc.telnet.Options.ANSI256 = false
			player.send("256 color mode is now disabled.")
		} else {
			player.desc.telnet.Options.ANSI256 = true
			player.send("[0882[094[0945[1006 [064c[028o[029l[030o[024r [018m[054o[090d[089e[088.[x is now enabled.")
		}
		return
	}
	if strings.EqualFold("UTF", input) {
		if telnet.Options.UTF {
			player.desc.telnet.Options.UTF = false
			player.send("UTF mode disabled.")
		} else {
			player.desc.telnet.Options.UTF = true
			player.send("UTF mode enabled.")
		}
		player.send("Character map test:")
		player.sendTestString()
		return
	}
	if strings.EqualFold("supga", input) {
		if telnet.Options.SuppressGoAhead {
			player.desc.telnet.Options.SuppressGoAhead = false
			player.send("SUPGA mode disabled.")
		} else {
			player.desc.telnet.Options.SuppressGoAhead = true
			player.send("SUPGA mode enabled.")
		}
		return
	}
	if strings.EqualFold("charmaps", input) {
		player.send("Character map list:")
		var buf string
		var count int
		for cname := range charsetList {
			count++
			buf = buf + fmt.Sprintf("%18v", cname)
			if count%4 == 0 {
				buf = buf + NEWLINE
			}
		}
		player.send(buf)
		player.send("To enable UTF, type 'telnet utf'")
		return
	}
	for cname, cset := range charsetList {
		if strings.EqualFold(input, cname) {
			player.desc.telnet.Charset = cname
			player.desc.telnet.charMap = cset
			player.desc.telnet.Options.UTF = false
			player.send("Your character map has been changed to: %v", cname)
			player.send("Character set test:")
			player.sendTestString()
			return
		}
	}
	player.send("That isn't a valid character map.")
}

func sendCmd(conn net.Conn, command, option byte) error {
	_, err := conn.Write([]byte{TermCmd_IAC, command, option})
	if err != nil {
		return err
	}
	return nil
}

func (desc *descData) sendSubSeq(data string, args ...byte) error {
	buf := []byte{TermCmd_IAC, TermCmd_SB}
	buf = append(buf, args...)
	if data != "" {
		buf = append(buf, []byte(data)...)
	}
	buf = append(buf, []byte{TermCmd_IAC, TermCmd_SE}...)
	dlen, err := desc.conn.Write(buf)
	if err != nil || dlen != len(buf) {
		return err
	}
	return nil
}

func (desc *descData) inputFull() {
	desc.sendln(warnBuf)
	buf := "Input buffer full! Stop spamming! Closing connection..."
	desc.sendln(buf)
	critLog("#%v: ERROR: %v: %v", desc.id, desc.ip, buf)
	desc.kill()
}

func (desc *descData) readByte() (byte, error) {
	data, err := desc.reader.ReadByte()
	if err != nil {
		descLock.Lock()
		desc.kill()
		descLock.Unlock()
		return 0, err
	}
	return data, nil
}
