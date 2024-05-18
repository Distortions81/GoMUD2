package main

import (
	"crypto/tls"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	KEYFILE  = "privkey.pem"
	CERTFILE = "fullchain.pem"
)

var (
	listener    *net.TCPListener
	listenerTLS net.Listener
)

func setupListenerTLS() {
	if *noTLS {
		return
	}

	cert, err := tls.LoadX509KeyPair(DATA_DIR+CERTFILE, DATA_DIR+KEYFILE)
	if err != nil {
		critLog("Error: %v", err.Error())
		critLog("Error loading TLS certificates: %v & %v in %v directory... TLS port not opened.", KEYFILE, CERTFILE, DATA_DIR)
		critLog("Use makeTestCert.sh, or letsencrypt if you have a domain name.")
		return
	}

	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	addr, err := net.ResolveTCPAddr("tcp4", *bindIP+":"+strconv.Itoa(*portTLS))
	if err != nil {
		critLog("Unable to resolve %v:%v: Error: %v", *bindIP, *portTLS, err)
		os.Exit(1)
	}

	listenerTLS, err = tls.Listen("tcp4", addr.String(), tlsCfg)
	if err != nil {
		critLog("Unable to listen at %v:%v. Error: %v", *bindIP, *portTLS, err)
		os.Exit(1)
	}

	mudLog("TLS listener online at: %s", addr.String())
}

func setupListener() {
	//Find network
	addr, err := net.ResolveTCPAddr("tcp4", *bindIP+":"+strconv.Itoa(*port))
	if err != nil {
		critLog("Unable to resolve %v:%v: Error: %v", *bindIP, *port, err)
		os.Exit(1)
	}

	//Open listener
	listener, err = net.ListenTCP("tcp4", addr)
	if err != nil {
		critLog("Unable to listen on port %v:%v. Error: %v", *bindIP, *port, err)
		os.Exit(1)
	}

	//Print listener
	mudLog("TCP listener online at: %s", addr.String())
}

func waitNewConnectionSSL() {

	if !*noTLS && listenerTLS != nil {

		for serverState.Load() == SERVER_RUNNING {

			conn, err := listenerTLS.Accept()
			if err != nil {
				critLog("Listener error: %v -- exiting loop", err)
				break
			}

			go handleDesc(conn, true)
			time.Sleep(CONNECT_THROTTLE)
		}

		listenerTLS.Close()
	}
}

func waitNewConnection() {

	for serverState.Load() == SERVER_RUNNING && listener != nil {

		conn, err := listener.Accept()
		if err != nil {
			critLog("Listener error: %v -- exiting loop", err)
			break
		}

		go handleDesc(conn, false)
		time.Sleep(CONNECT_THROTTLE)
	}

	listener.Close()
}
