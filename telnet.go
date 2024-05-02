package main

func (desc *descData) sendTelnetCmds() {
	desc.sendCmd(TermCmd_DO, TermOpt_SUP_GOAHEAD)
	desc.sendCmd(TermCmd_DO, TermOpt_TERMINAL_TYPE)
	desc.sendCmd(TermCmd_WILL, TermOpt_CHARSET)
	desc.sendCmd(TermCmd_WILL, TermOpt_SUP_GOAHEAD)
}

func (desc *descData) sendCmd(command, option byte) error {
	dlen, err := desc.conn.Write([]byte{TermCmd_IAC, command, option})
	if err != nil || dlen != 3 {
		//errLog("#%v: %v: command send failed (connection lost)", desc.id, desc.cAddr)
		return err
	}

	//errLog("#%v: Sent: %v %v", desc.id, TermCmd2Txt[int(command)], TermOpt2TXT[int(option)])
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

	if len(args) > 1 {
		//errLog("#%v: Sent sub: %v %v %d", desc.id, data, TermOpt2TXT[int(args[0])], args[1])
	}

	return nil
}

func (desc *descData) inputFull() {
	buf := "Input buffer full! Closing connection..."
	desc.sendln("\r\n%v", buf)
	critLog("#%v: ERROR: %v: %v", desc.id, desc.cAddr, buf)
	desc.close()
}

func (desc *descData) readByte() (byte, error) {
	data, err := desc.reader.ReadByte()
	if err != nil {
		errLog("#%v: %v: Connection closed by server.", desc.id, desc.cAddr)
		return 0, err
	}
	return data, nil
}

func (desc *descData) close() {
	if desc == nil {
		return
	}
	desc.state = CON_DISCONNECTED
	if desc.account != nil {
		desc.account.tempString = ""
	}
	desc.valid = false
	if desc.conn != nil {
		desc.conn.Close()
	}
}
