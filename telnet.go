package main

import "net"

func sendTelnetCmds(conn net.Conn) {
	sendCmd(conn, TermCmd_DO, TermOpt_SUP_GOAHEAD)
	sendCmd(conn, TermCmd_DO, TermOpt_TERMINAL_TYPE)
	sendCmd(conn, TermCmd_WILL, TermOpt_CHARSET)
	sendCmd(conn, TermCmd_WILL, TermOpt_SUP_GOAHEAD)
}

func sendCmd(conn net.Conn, command, option byte) error {
	_, err := conn.Write([]byte{TermCmd_IAC, command, option})
	if err != nil {
		return err
	}
	return nil
}

func (desc *descData) sendSub(data string, args ...byte) error {
	buf := []byte{TermCmd_IAC, TermCmd_SB}
	buf = append(buf, args...)
	if data != "" {
		buf = append(buf, []byte(data)...)
	}
	buf = append(buf, []byte{TermCmd_IAC, TermCmd_SE}...)
	dlen, err := desc.conn.Write(buf)
	if err != nil || dlen != len(buf) {
		//errLog("#%v: %v: sub send failed (connection lost)", desc.id, desc.cAddr)
		return err
	}

	/*
		if len(args) > 1 {
			errLog("#%v: Sent sub: %v %v %d", desc.id, data, TermOpt2TXT[int(args[0])], args[1])
		}
	*/

	return nil
}

func (desc *descData) inputFull() {
	desc.send(warnBuf)
	buf := "Input buffer full! Stop spamming. Closing connection..."
	desc.sendln(buf)
	critLog("#%v: ERROR: %v: %v", desc.id, desc.cAddr, buf)
	desc.valid = false
	desc.state = CON_DISCONNECTED
}

func (desc *descData) readByte() (byte, error) {
	data, err := desc.reader.ReadByte()
	if err != nil {
		//errLog("#%v: %v: Connection closed by server.", desc.id, desc.cAddr)
		descLock.Lock()
		desc.valid = false
		desc.state = CON_DISCONNECTED
		descLock.Unlock()
		return 0, err
	}
	return data, nil
}

func (desc *descData) close() {
	if desc == nil {
		return
	}
	desc.state = CON_DISCONNECTED
	desc.valid = false
	if desc.character != nil {
		desc.character.desc = nil
	}
}
