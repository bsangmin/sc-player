package main

import (
	tm "github.com/buger/goterm"
	con "github.com/containerd/console"
	"time"
)

func console() {
	c := con.Current()
	defer c.Reset()

	for {
		tm.Clear()

		tm.MoveCursor(2, 2)
		tm.Print(mapInfo.StringStatus())

		tm.MoveCursor(9, 2)
		txt := tm.Color(mapInfo.Name.ToString(), tm.GREEN)
		tm.Print(txt)

		tm.MoveCursor(2, 4)
		tm.Print(tm.Background("   ", tm.GREEN), ": Host", "  ", tm.Background("   ", tm.BLUE), ": Me", "  ", tm.Background("   ", tm.RED), ": Out")

		idx := 6
		for _, p := range players {
			if p == nil || p.Name == nil || p.Batcode == nil || p.IP == nil {
				continue
			}

			tm.MoveCursor(2, idx)
			if p.Num == 0 {
				txt := tm.Color("0", tm.BLACK)
				txt = tm.Background(txt, tm.GREEN)
				tm.Print(txt)

			} else {
				tm.Print(p.Num)
			}

			tm.MoveCursor(7, idx)
			if p.Me {
				txt := tm.Color(p.Name.ToString(), tm.BLACK)
				txt = tm.Background(txt, tm.BLUE)
				tm.Print(txt)

			} else if p.Out {
				txt := tm.Color(p.Name.ToString(), tm.BLACK)
				txt = tm.Background(txt, tm.RED)
				tm.Print(txt)

			} else {
				tm.Print(p.Name.ToString())
			}

			tm.MoveCursor(25, idx)
			tm.Print(p.Batcode.ToString())
			tm.MoveCursor(55, idx)
			tm.Print(p.IP.String())

			idx++
		}
		tm.MoveCursor(1, idx)
		tm.Flush()
		time.Sleep(500 * time.Millisecond)
	}
}
