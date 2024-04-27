package main

import "fmt"

func (desc *descData) send(format string, args ...any) (fail bool) {

	var data string
	if args != nil {
		data = fmt.Sprintf(format, args...)
	} else {
		data = format
	}

	dlen := len(data)
	l, err := desc.conn.Write([]byte(data))

	if err != nil || dlen != l {
		mudLog("#%v: %v: write failed (connection lost)", desc.id, desc.conn.RemoteAddr().String())
		desc.state = CON_DISCONNECTED
		desc.conn.Close()
		return true
	}

	return false
}
