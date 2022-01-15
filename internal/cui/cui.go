package cui

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/sinnershiki/pingo/pkg/ping"
)

func updateResultView(g *gocui.Gui) error {
	x0, _, x1, _, err := g.ViewPosition("result")
	if err != nil {
		return err
	}
	msg := "[Num] <Dest IP>: Average TTL, Reached ping rate, Unreached ping count\n"
	msg += strings.Repeat("-", x1-x0) + "\n"
	for i, p := range ping.Pingers {
		sum := float64(0)
		for _, rtt := range p.Rtts {
			sum += float64(rtt.Microseconds())
		}

		ip := p.IP.String()
		rate := float64(p.ReceivedCount) / (float64(p.ReceivedCount) + float64(p.ErrorCount)) * 100
		avg_rtt := float64(sum) / float64(len(p.Rtts)) / 1000
		msg += fmt.Sprintf("[%d] %s: avg_rtt=%gms, rate=%.2f, received=%d, error=%d\n", i+1, ip, avg_rtt, rate, p.ReceivedCount, p.ErrorCount)
	}

	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("result")
		if err != nil {
			return err
		}
		v.Clear()
		fmt.Fprintln(v, msg)
		return nil
	})

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func inputedIp(g *gocui.Gui, v *gocui.View) error {
	ipStr := g.CurrentView().BufferLines()[0]
	ip := net.ParseIP(ipStr)

	if _, err := g.SetCurrentView("result"); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.DeleteView("inputIp"); err != nil {
		return err
	}

	p := ping.NewPinger(ip)
	p.PingV4()
	ping.Pingers = append(ping.Pingers, p)

	if err := updateResultView(g); err != nil {
		return err
	}

	return nil
}

func inputedResultNum(g *gocui.Gui, v *gocui.View) error {
	num, err := strconv.Atoi(g.CurrentView().BufferLines()[0])
	if err != nil {
		log.Panicln(err)
		return err
	}

	if _, err := g.SetCurrentView("result"); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.DeleteView("inputResultNum"); err != nil {
		return err
	}

	ping.RemoveByIndex(num - 1)

	if err := updateResultView(g); err != nil {
		return err
	}

	return nil
}

func quitInputView(g *gocui.Gui, v *gocui.View) error {
	viewname := g.CurrentView().Title

	if _, err := g.SetCurrentView("result"); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.DeleteView(viewname); err != nil {
		return err
	}

	return nil
}

func setInputIpView(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("inputIp", maxX/2-10, maxY/2-1, maxX/2+9, maxY/2+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Input IP here"
		v.Editable = true
	}

	// フォーカスを変更
	if _, err := g.SetCurrentView("inputIp"); err != nil {
		log.Panicln(err)
		return err
	}

	return nil
}

func setInputResultNumView(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("inputResultNum", maxX/2-10, maxY/2-1, maxX/2+9, maxY/2+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Input Result Number"
		v.Editable = true
	}

	// フォーカスを変更
	if _, err := g.SetCurrentView("inputResultNum"); err != nil {
		log.Panicln(err)
		return err
	}

	return nil
}

func InitKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.SetKeybinding("inputIp", gocui.KeyEnter, gocui.ModNone, inputedIp); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.SetKeybinding("inputIp", 'q', gocui.ModNone, quitInputView); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.SetKeybinding("inputResultNum", gocui.KeyEnter, gocui.ModNone, inputedResultNum); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.SetKeybinding("inputResultNum", 'q', gocui.ModNone, quitInputView); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.SetKeybinding("", 'n', gocui.ModNone, setInputIpView); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.SetKeybinding("", 'd', gocui.ModNone, setInputResultNumView); err != nil {
		log.Panicln(err)
		return err
	}
	return nil
}
