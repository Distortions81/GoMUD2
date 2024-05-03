package main

import (
	"net"
	"os"
	"strconv"
	"time"
)

const (
	SSL_PEM = "fullchain.pem"
	SSL_KEY = "privkey.pem"
)

var (
	listener *net.TCPListener
	//listenerTLS net.Listener
)

/*
func setupListenerTLS() {
	if *noTLS {
		return
	}
	//openssl ecparam -genkey -name prime256v1 -out server.key
	//openssl req -new -x509 -key server.key -out server.pem -days 3650

	cert, err := tls.LoadX509KeyPair(DATA_DIR+SSL_PEM, DATA_DIR+SSL_KEY)
	if err != nil {
		errLog("Error loading TLS certificates: %v & %v in %v directory... TLS port not opened.", SSL_PEM, SSL_KEY, DATA_DIR)
		errLog("Use makeTestCert.sh, or letsencrypt if you have a domain name.")
		return
	}

	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	addr, err := net.ResolveTCPAddr("tcp4", *bindIP+":"+strconv.Itoa(*portTLS))
	if err != nil {
		errLog("Unable to resolve %v:%v: Error: %v", *bindIP, *portTLS, err)
		os.Exit(1)
	}

	listenerTLS, err = tls.Listen("tcp4", addr.String(), tlsCfg)
	if err != nil {
		errLog("Unable to listen at %v:%v. Error: %v", *bindIP, *portTLS, err)
		os.Exit(1)
	}

	errLog("TLS listener online at: %s", addr.String())
}
*/

func setupListener() {
	//Find network
	addr, err := net.ResolveTCPAddr("tcp4", *bindIP+":"+strconv.Itoa(*port))
	if err != nil {
		errLog("Unable to resolve %v:%v: Error: %v", *bindIP, *port, err)
		os.Exit(1)
	}

	//Open listener
	listener, err = net.ListenTCP("tcp4", addr)
	if err != nil {
		errLog("Unable to listen on port %v:%v. Error: %v", *bindIP, *port, err)
		os.Exit(1)
	}

	//Print listener
	errLog("TCP listener online at: %s", addr.String())
}

/*
func waitNewConnectionSSL() {

	if !*noTLS && portTLS != nil && listenerTLS != nil {

		for serverState.Load() == SERVER_RUNNING {

			conn, err := listenerTLS.Accept()
			if err != nil {
				errLog("Listener error: %v -- exiting loop", err)
				break
			}

			go handleDesc(conn, true)
			time.Sleep(CONNECT_THROTTLE)
		}

		listenerTLS.Close()
	}
}
*/

func waitNewConnection() {

	for serverState.Load() == SERVER_RUNNING && listener != nil {

		conn, err := listener.Accept()
		if err != nil {
			errLog("Listener error: %v -- exiting loop", err)
			break
		}

		go handleDesc(conn, false)
		time.Sleep(CONNECT_THROTTLE)
	}

	listener.Close()
}
