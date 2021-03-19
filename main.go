package main

import (
	"flag"
	"io/ioutil"
	"log"
	"runtime"
)

// VERSION is version
var VERSION = "DEV"

var typeFlag = flag.String("type", "gui", "type: gui, console, log")

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	flag.Parse()

	if *typeFlag == "gui" {
		initGui()
	}
}

func main() {
	log.Println("SC player VER:", VERSION)

	runtime.LockOSThread()
	done := make(chan bool)

	go getPackets()

	switch *typeFlag {
	case "gui":
		log.SetOutput(ioutil.Discard)
		gui()
	case "console":
		log.SetOutput(ioutil.Discard)
		console()
	default:
		<-done
	}
}
