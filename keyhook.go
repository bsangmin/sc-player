package main

import (
	hook "github.com/robotn/gohook"
	"log"
)

var overlayShow = true

func keyHook() {
	hook.Register(hook.KeyDown, []string{"l", "ctrl", "shift"}, func(e hook.Event) {
		overlayShow = !overlayShow
		log.Println("show overlay:", overlayShow)
	})
	s := hook.Start()
	<-hook.Process(s)
}
