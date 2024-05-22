package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/tklauser/numcpus"
)

const (
	DEFAULT_PORT     = 7777
	DEFAULT_TLS_PORT = DEFAULT_PORT + 1

	SERVER_BOOTING = iota
	SERVER_RUNNING
	SERVER_SHUTDOWN
)

var (
	bootTime time.Time

	signalHandle chan os.Signal

	port          *int
	portTLS       *int
	noTLS         *bool
	makeTestFiles *bool
	bindIP        *string
	serverState   atomic.Int32
	numThreads    int
)

func main() {
	serverState.Store(SERVER_BOOTING)
	bootTime = time.Now()

	port = flag.Int("port", DEFAULT_PORT, "port")
	portTLS = flag.Int("portTLS", DEFAULT_TLS_PORT, "TLS Port")
	noTLS = flag.Bool("noSSL", true, "disable TLS listener")
	bindIP = flag.String("bindIP", "localhost", "Bind to a specific IP.")
	makeTestFiles = flag.Bool("fileBootstrap", false, "Create simple example area and help files.")
	flag.Parse()

	//Make sure all directories we need are created
	for _, newDir := range makeDirs {
		err := os.Mkdir(newDir, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			log.Printf("Unable to create directory: %v: %v", newDir, err)
			os.Exit(1)
		}
	}

	var err error
	numThreads, err = numcpus.GetOnline()
	if err != nil {
		numThreads = runtime.NumCPU()
	}

	startLogs()
	loadMudID()

	if *makeTestFiles {
		makeTestArea()
		saveAllAreas(true)

		makeTestHelp()
		saveHelps()

		critLog("Bootstrap files created.")
	}

	loadAllAreas()
	readTextFiles()

	loadHelps()
	saveHelps()
	readDisables()

	setupListener()
	setupListenerTLS()

	loadAccountIndex()

	serverState.Store(SERVER_RUNNING)

	go waitNewConnection()
	go waitNewConnectionSSL()
	go mainLoop()
	go hasherDaemon()

	//After starting loops, wait here for process signals
	signalHandle = make(chan os.Signal, 1)

	signal.Notify(signalHandle, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalHandle

	saveCharacters(true)
	//saveAllAreas(true)
	serverState.Store(SERVER_SHUTDOWN)
	time.Sleep(time.Second)

	//Handle shutdown here
	closeLogs()
}
