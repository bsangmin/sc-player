package main

import (
	"flag"
	"io/ioutil"
	"log"
	"runtime"
)

// VERSION is version

var PROG = "SC Player"
var VERSION = "DEV"

var typeFlag = flag.String("type", "overlay", "type: gui, console, overlay")

func init() {
	runtime.LockOSThread()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	flag.Parse()

	if *typeFlag == "gui" || *typeFlag == "overlay" {
		initGui()
	}
}

func main() {
	if VERSION != "DEV" {
		log.SetOutput(ioutil.Discard)
	}

	// log.Println("SC player VER:", VERSION)
	log.Printf("%s %s\n", PROG, VERSION)

	done := make(chan bool)

	go getPackets()

	switch *typeFlag {
	case "gui":
		gui()
	case "overlay":
		overlay()
	case "console":
		console()
	default:
		<-done
	}
}
