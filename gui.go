package main

import (
	"fmt"
	"log"
	"syscall"

	"unsafe"

	"path/filepath"
	"runtime"

	"github.com/bsangmin/sc-player/location"
	lwin "github.com/lxn/win"
	"github.com/sciter-sdk/go-sciter"
	swin "github.com/sciter-sdk/go-sciter/window"
)

var window *swin.Window

const (
	WIDTH  = 670
	HEIGHT = 600
)

func getScreenSize() (int, int) {
	hDC := lwin.GetDC(0)
	defer lwin.ReleaseDC(0, hDC)
	width := int(lwin.GetDeviceCaps(hDC, lwin.HORZRES))
	height := int(lwin.GetDeviceCaps(hDC, lwin.VERTRES))

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

	x := int((sw - WIDTH) / 2)
	y := int((sh - HEIGHT) / 2)

	if x < 0 {
		x = 100
	}
	if y < 0 {
		y = 100
	}

	rect := sciter.NewRect(y, x, WIDTH, HEIGHT)
	w, windowCreationErr := swin.New(sciter.SW_MAIN|sciter.SW_TOOL|sciter.SW_ALPHA, rect)
	// w, windowCreationErr := swin.New(sciter.SW_MAIN|sciter.SW_GLASSY|sciter.SW_CONTROLS, rect)
	w.SetTitle(PROG)

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

	window = w

}

func getPos(hwnd lwin.HWND, right bool, bottom bool) (x int32, y int32) {
	var wRect, cRect lwin.RECT
	lwin.GetWindowRect(hwnd, &wRect)
	lwin.GetClientRect(hwnd, &cRect)

	margin := int32(20)

	xDir := int32(1)
	yDir := int32(1)

	if right {
		x = wRect.Right - WIDTH
		xDir = -1
	} else {
		x = wRect.Left
	}

	if bottom {
		y = wRect.Bottom - HEIGHT
		yDir = -1
	} else {
		y = wRect.Top
	}

	// window mode
	if wRect.Right != cRect.Right {
		x += 10 * xDir // for win10
	}

	if wRect.Bottom != cRect.Bottom && !bottom {
		y += 35 * yDir // for win10
	}

	x += margin * xDir
	y += margin * yDir

	return
}

func setEventHandler(w *swin.Window) {
	var layoutHwnd lwin.HWND
	ptr, _ := syscall.UTF16PtrFromString("Brood War")

	w.DefineFunction("init", func(args ...*sciter.Value) *sciter.Value {
		code := `$(#players).clear(); 
		for(var i=0; i<maxPlayers; i++){
			$(#players).$append(<div.hide-player.player#p{i}>
			<div.num><div.data></div></div>
			<div.clickable.name><div.data></div></div>
			<div.clickable.bat><div.data></div></div>
			<div.clickable.ip><div.data></div></div>
			</div>);
		};
		var ver = $(body).$append(<div#version>%s %s</div>)`
		w.Eval(fmt.Sprintf(code, PROG, VERSION))

		return sciter.NullValue()
	})

	w.DefineFunction("where", func(args ...*sciter.Value) *sciter.Value {
		go func(data *sciter.Value, callback *sciter.Value) {
			if VERSION != "DEV" {
				callback.Invoke(sciter.NullValue(), "[Native Script]", sciter.NewValue(""))
				return
			}

			loc, err := location.FindLocation(data.String())
			if err != nil {
				log.Println(err)
				callback.Invoke(sciter.NullValue(), "[Native Script]", sciter.NewValue(err.Error()))
			} else {
				callback.Invoke(sciter.NullValue(), "[Native Script]", sciter.NewValue(loc))
			}

		}(args[0].Clone(), args[1].Clone())

		return sciter.NullValue()
	})

	w.DefineFunction("refresh", func(args ...*sciter.Value) *sciter.Value {

		if !overlayShow {
			return sciter.NewValue(0x0)
		}

		hwnd := lwin.FindWindow(nil, ptr)
		if hwnd == 0 {
			return sciter.NewValue(0x0)
		}

		if layoutHwnd == 0 {
			layoutHwnd = lwin.HWND(unsafe.Pointer(w.GetHwnd()))
		}

		payload := sciter.NewValue()

		currentWinHwnd := lwin.GetForegroundWindow()
		if layoutHwnd == currentWinHwnd {
			payload.Append(0x2)

		} else if hwnd == currentWinHwnd {
			x, y := getPos(hwnd, true, true)

			payload.Append(0x1)
			payload.Append(x)
			payload.Append(y)

		} else {
			payload.Append(0x0)
		}

		return payload
	})

	w.DefineFunction("plz", func(args ...*sciter.Value) *sciter.Value {
		payload := sciter.NewValue()
		playerSlice := sciter.NewValue()

		payload.Set("title", mapInfo.StringName())
		payload.Set("scode", mapInfo.Status)
		payload.Set("status", mapInfo.StringStatus())

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

			if VERSION == "DEV" {
				player.Set("ip", p.IP.String())
			} else {
				player.Set("ip", p.HiddenIP().String())
			}

			playerSlice.Append(player)
		}

		// playerSlice = testUsers(12)

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

func testUsers(count int) *sciter.Value {
	playerSlice := sciter.NewValue()

	for i := 0; i < count; i++ {
		player := sciter.NewValue()

		player.Set("num", i)
		player.Set("me", false)
		player.Set("out", false)

		player.Set("name", "WWWWWWWWWWWWWWW")
		player.Set("bat", "무지막지한무지#123456")
		player.Set("ip", "106.255.140.180")

		playerSlice.Append(player)
	}
	return playerSlice
}

func gui() {
	window.Show()
	window.Run()
}
