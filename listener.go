package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"time"
)

const (
	SSL_PEM = "server.pem"
	SSL_KEY = "server.key"
)

var (
	listener    *net.TCPListener
	listenerTLS net.Listener
)

func setupListenerTLS() {
	if *noTLS {
		return
	}
	//openssl ecparam -genkey -name prime256v1 -out server.key
	//openssl req -new -x509 -key server.key -out server.pem -days 3650

	cert, err := tls.LoadX509KeyPair(DATA_DIR+SSL_PEM, DATA_DIR+SSL_KEY)
	if err != nil {
		errLog("Error loading SSL certificate, SSL port not opened.")
		errLog("How to make cert: (put in data directory)")
		errLog("openssl ecparam -genkey -name prime256v1 -out server.key")
		errLog("openssl req -new -x509 -key server.key -out server.pem -days 3650")
		return
	}

	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	/*Open Listener*/
	listenerTLS, err = tls.Listen("tcp4", strconv.Itoa(DEFAULT_TLS_PORT), tlsCfg)
	if err != nil {

	}

	/*Print Connection*/
	buf := fmt.Sprintf("SSL listener online at: %s", DEFAULT_TLS_PORT)
	errLog(buf)
}

func setupListener() {
	/*Find Network*/
	addr, err := net.ResolveTCPAddr("tcp4", DEFAULT_PORT)
	support.CheckError("main: resolveTCP", err, ERROR_FATAL)

	/*Open Listener*/
	listener, err := net.ListenTCP("tcp4", addr)
	glob.ServerListener = listener
	support.CheckError("main: ListenTCP", err, ERROR_FATAL)

	/*Print Connection*/
	buf := fmt.Sprintf("TCP listener online at: %s", addr.String())
	mlog.Write(buf)
}

func waitNewConnectionSSL() {

	if glob.ServerListenerSSL != nil {

		for glob.ServerState == SERVER_RUNNING {

			time.Sleep(CONNECT_THROTTLE_MS * time.Millisecond)
			desc, err := glob.ServerListenerSSL.Accept()
			support.AddNetDesc()

			/* If there is a connection flood, sleep listeners */
			if err != nil || support.CheckNetDesc() {
				time.Sleep(5 * time.Second)
				desc.Close()
				support.RemoveNetDesc()
			} else {

				desc.Write([]byte(
					LICENSE + support.TextFiles["greet"] +
						"(SSL Encryption Enabled!)\n(Type NEW to create character) Name:"))
				time.Sleep(CONNECT_THROTTLE_MS * time.Millisecond)
				support.NewDescriptor(desc, true)
			}

		}

		glob.ServerListenerSSL.Close()
	}
}

func waitNewConnection() {

	for glob.ServerState == SERVER_RUNNING {

		time.Sleep(CONNECT_THROTTLE_MS * time.Millisecond)
		desc, err := glob.ServerListener.Accept()
		support.AddNetDesc()
		time.Sleep(CONNECT_THROTTLE_MS * time.Millisecond)

		/* If there is a connection flood, sleep listeners */
		if err != nil || support.CheckNetDesc() {
			time.Sleep(5 * time.Second)
			desc.Close()
			support.RemoveNetDesc()
		} else {

			_, err = desc.Write([]byte(
				LICENSE + support.TextFiles["greet"] +
					"(ENCRYPTION NOT ENABLED!)\n(Type NEW to create character) Name:"))

			time.Sleep(CONNECT_THROTTLE_MS * time.Millisecond)
			support.NewDescriptor(desc, false)
		}
	}

	glob.ServerListener.Close()
}
