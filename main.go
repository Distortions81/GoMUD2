package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
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

	port           *int
	portTLS        *int
	noTLS          *bool
	makeTestFiles  *bool
	instantRespond *bool
	perfProfile    *bool
	bindIP         *string
	serverState    int
	numThreads     int
)

func main() {

	port = flag.Int("port", DEFAULT_PORT, "port")
	portTLS = flag.Int("portTLS", DEFAULT_TLS_PORT, "TLS Port")
	noTLS = flag.Bool("noSSL", true, "disable TLS listener")
	bindIP = flag.String("bindIP", "localhost", "Bind to a specific IP.")
	makeTestFiles = flag.Bool("fileBootstrap", false, "Create simple example area and help files.")
	instantRespond = flag.Bool("instantRespond", true, "Respond to commands instantly, instead of once per pulse.")
	perfProfile = flag.Bool("perfProfile", false, "Launch and performance profile.")
	flag.Parse()

	if *perfProfile {
		f, _ := os.Create(DATA_DIR + "perf.prof")
		pprof.StartCPUProfile(f)

		go func() {
			time.Sleep(time.Minute)
			pprof.StopCPUProfile()
			os.Exit(0)
		}()
	}

	mudMain()
}

func mudMain() {
	serverState = SERVER_BOOTING
	bootTime = time.Now().UTC()
	//Make sure all directories we need are created
	for _, newDir := range makeDirs {
		err := os.Mkdir(newDir, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			log.Printf("Unable to create directory: %v: %v", newDir, err)
			os.Exit(1)
		}
	}

	GetOsTimeZones()

	var err error
	numThreads, err = numcpus.GetOnline()
	if err != nil {
		numThreads = runtime.NumCPU()
	}

	startLogs()
	loadMudID()
	loadMudStats()

	if *makeTestFiles {
		makeTestArea()
		saveAllAreas(true)

		makeTestHelp()
		saveHelps()
		saveNotes(true)

		critLog("Bootstrap files created.")
	}

	servSet = loadSettings()
	loadAllAreas()
	loadTextFiles()
	loadEmojiHelp()
	updateFontList()

	loadHelps()
	xcolorHelp()

	saveHelps()
	loadBlocked()
	loadDisables()
	loadNotes()
	loadBugs()
	saveNotes(true)

	setupListener()
	setupListenerTLS()

	loadAccountIndex()

	serverState = SERVER_RUNNING

	go waitNewConnection()
	go waitNewConnectionSSL()
	go hasherDaemon()
	go mainLoop()

	//After starting loops, wait here for process signals
	signalHandle = make(chan os.Signal, 1)

	signal.Notify(signalHandle, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalHandle

	sendToAll("--> Saving areas <--")
	saveAllAreas(false)
	sendToAll("--> Saving players <--")
	saveAllCharacters(false)
	sendToAll("--> Server rebooting <--")
	time.Sleep(time.Second)
	serverState = SERVER_SHUTDOWN

	//Handle shutdown here
	closeLogs()
}
