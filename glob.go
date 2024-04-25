package main

import (
	"os"
	"time"
)

var (
	bootTime time.Time

	signalHandle chan os.Signal

	port, portTLS *int
	noTLS         *bool
	bindIP        *string
)
