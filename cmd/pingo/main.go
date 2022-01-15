package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/sinnershiki/pingo/internal/cui"
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Println(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := cui.InitKeybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func init() {
	logFile := "./pingo.log"
	logfile, _ := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	log.SetOutput(logfile)
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("result", 0, 0, maxX-1, maxY*2/3-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		if _, err := g.SetCurrentView("result"); err != nil {
			return err
		}

		v.Title = "Pingo"

		x, _ := v.Size()
		msg := "[Num] <Dest IP>: Average TTL, Reached ping rate, Unreached ping count\n"
		msg += strings.Repeat("-", x) + "\n"
		fmt.Fprintln(v, msg)
	}

	if v, err := g.SetView("console", 0, maxY*2/3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Console"
		//TODO: hもつけてヘルプみたいなの出したい
		fmt.Fprintln(v, "input: n, d")
	}

	return nil
}
