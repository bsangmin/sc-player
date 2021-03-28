package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/getlantern/systray"
)

func overlay() {
	go keyHook()
	systray.Run(onReady, onExit)

	gui()
}

func onReady() {
	getIcon := func(s string) []byte {
		b, err := ioutil.ReadFile(s)
		if err != nil {
			log.Fatalln("faild to load icon:", err)
		}
		return b
	}

	systray.SetIcon(getIcon("./res/i32.ico"))
	systray.SetTitle(PROG)
	systray.SetTooltip(fmt.Sprintf("%s running...", PROG))

	mQuit := systray.AddMenuItem("종료", "프로그램 종료")

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	os.Exit(0)
}
