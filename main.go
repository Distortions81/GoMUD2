package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	DEFAULT_PORT     = 7777
	DEFAULT_TLS_PORT = DEFAULT_PORT + 1
)

var (
	bootTime time.Time

	signalHandle chan os.Signal

	port, portTLS *int
	noTLS         *bool
	bindIP        *string
	serverState   int
)

func main() {
	bootTime = time.Now()

	port = flag.Int("port", DEFAULT_PORT, "port")
	portTLS = flag.Int("portTLS", DEFAULT_TLS_PORT, "TLS Port")
	noTLS = flag.Bool("noSSL", false, "disable TLS listener")
	bindIP = flag.String("bindIP", "localhost", "Bind to a specific IP.")
	flag.Parse()

	//Make sure all directories we need are created
	for _, newDir := range makeDirs {
		err := os.Mkdir(newDir, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			log.Printf("Unable to create directory: %v: %v", newDir, err)
			os.Exit(1)
		}
	}

	startLogs()

	setupListener()
	setupListenerTLS()

	serverState = SERVER_RUNNING
	go waitNewConnection()
	go waitNewConnectionSSL()

	go mainLoop()

	//After starting loops, wait here for process signals
	signalHandle = make(chan os.Signal, 1)

	signal.Notify(signalHandle, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalHandle

	//Handle shutdown here
	closeLogs()
}
