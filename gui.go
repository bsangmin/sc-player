package main

import (
	"log"

	"github.com/sciter-sdk/go-sciter"
	"github.com/sciter-sdk/go-sciter/window"

	"path/filepath"
	"runtime"

	ww "github.com/lxn/win"
)

var win *window.Window

func getScreenSize() (int, int) {
	hDC := ww.GetDC(0)
	defer ww.ReleaseDC(0, hDC)
	width := int(ww.GetDeviceCaps(hDC, ww.HORZRES))
	height := int(ww.GetDeviceCaps(hDC, ww.VERTRES))

	return width, height
}

func initGui() {
	log.Println("init sciter")

	if runtime.GOARCH == "amd64" {
		sciter.SetDLL("./res/x64/sciter.dll")
	} else {
		sciter.SetDLL("./res/x86/sciter.dll")
	}

	sw, sh := getScreenSize()
	ww, wh := 670, 600

	x := int((sw - ww) / 2)
	y := int((sh - wh) / 2)

	if x < 0 {
		x = 100
	}
	if y < 0 {
		y = 100
	}

	rect := sciter.NewRect(y, x, ww, wh)
	w, windowCreationErr := window.New(sciter.SW_MAIN|sciter.SW_GLASSY|sciter.SW_CONTROLS, rect)

	if windowCreationErr != nil {
		log.Fatalf("Could not create sciter window : %s",
			windowCreationErr.Error())
		return
	}

	fullpath, err := filepath.Abs("./res/main.html")
	if err != nil {
		log.Fatalf("Could not get fullpath main.html")
		return
	}

	if err := w.LoadFile(fullpath); err != nil {
		log.Fatalf("Could not load ui file : %s",
			err.Error())
	}

	setEventHandler(w)

	win = w

}

func setEventHandler(w *window.Window) {
	w.DefineFunction("plz", func(args ...*sciter.Value) *sciter.Value {

		payload := sciter.NewValue()
		playerSlice := sciter.NewValue()

		payload.Set("ver", VERSION)
		payload.Set("title", mapInfo.StringName())
		payload.Set("scode", mapInfo.Status)
		payload.Set("status", mapInfo.StringStatus())

		// players[0] = &structures.Player{
		// 	Num: 0, Me: true, Out: false, Name: []byte("WWWWWWWWWWWWWWW"), Batcode: []byte("무지막지한무지#123456"), IP: net.IPv4(255, 255, 255, 255),
		// }

		for _, p := range players {
			if p == nil || p.Name == nil || p.Batcode == nil || p.IP == nil {
				continue
			}

			player := sciter.NewValue()

			player.Set("num", int(p.Num))
			player.Set("me", p.Me)
			player.Set("out", p.Out)

			player.Set("name", p.Name.ToString())
			player.Set("bat", p.Batcode.ToString())
			player.Set("ip", p.IP.String())

			playerSlice.Append(player)
		}

		payload.Set("players", playerSlice)

		return payload
	})

	w.DefineFunction("glog", func(args ...*sciter.Value) *sciter.Value {
		for idx, arg := range args {
			log.Println("[glog] idx:", idx, "val: ", arg)
		}
		return sciter.NullValue()
	})
}

func gui() {
	win.Show()
	win.Run()
}
