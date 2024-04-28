package main

import "golang.org/x/text/encoding/charmap"

func sendTelnetCmds(desc *descData) {
	//desc.sendCmd(TermCmd_DO, TermOpt_SUP_GOAHEAD)
	desc.sendCmd(TermCmd_DO, TermOpt_TERMINAL_TYPE)
	desc.sendCmd(TermCmd_WILL, TermOpt_CHARSET)
	//desc.sendCmd(TermCmd_WILL, TermOpt_SUP_GOAHEAD)
}

func (desc *descData) sendCmd(command, option byte) error {
	dlen, err := desc.conn.Write([]byte{TermCmd_IAC, command, option})
	if err != nil || dlen != 3 {
		mudLog("#%v: %v: command send failed (connection lost)", desc.id, desc.cAddr)
		return err
	}

	errLog("#%v: Sent: %v %v", desc.id, TermCmd2Txt[int(command)], TermOpt2TXT[int(option)])
	return nil
}

func (desc *descData) sendEOR() error {
	dlen, err := desc.conn.Write([]byte{TermOpt_END_OF_RECORD})
	if err != nil || dlen != 1 {
		mudLog("#%v: %v: EOR send failed (connection lost)", desc.id, desc.cAddr)
		return err
	}
	errLog("#%v: Sent: %v", desc.id, TermOpt2TXT[int(TermOpt_END_OF_RECORD)])
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
		mudLog("#%v: %v: sub send failed (connection lost)", desc.id, desc.cAddr)
		return err
	}

	if len(args) > 1 {
		errLog("#%v: Sent sub: %v %v %d", desc.id, data, TermOpt2TXT[int(args[0])], args[1])
	}

	return nil
}

func (desc *descData) inputFull() {
	buf := "Input buffer full! Closing connection..."
	desc.send("\r\n%v\r\n", buf)
	mudLog("#%v: ERROR: %v: %v", desc.id, desc.cAddr, buf)
}

func (desc *descData) readByte() (byte, error) {
	data, err := desc.reader.ReadByte()
	if err != nil {
		mudLog("#%v: %v: Connection closed by server.", desc.id, desc.cAddr)
		return 0, err
	}
	if !desc.telnet.utf {
		data = convertByte(charmap.ISO8859_1, data)
	}
	return data, nil
}

func (desc *descData) close() {
	if desc == nil {
		return
	}
	desc.state = CON_DISCONNECTED
	if desc.conn != nil {
		desc.conn.Close()
	}
}
