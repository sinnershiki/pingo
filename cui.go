package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/jroimartin/gocui"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const IPV4_ICMP = 1

type PingResult struct {
	IP            net.IP
	ErrorCount    int
	ReceivedCount int
	ttls          []time.Duration
}

func main() {
	var results []PingResult
	localip := net.ParseIP("127.0.0.1")
	result := pingIpv4(localip, 5)
	results = append(results, result)

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		fmt.Println(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := initKeybindings(g); err != nil {
		log.Panicln(err)
	}

	updateResultView(g, results)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("result", 0, 0, maxX-1, maxY*2/3-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Pingo"
		fmt.Fprintln(v, "show results")
	}

	if v, err := g.SetView("console", 0, maxY*2/3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if _, err := g.SetCurrentView("console"); err != nil {
			return err
		}
		v.Title = "Console"
		fmt.Fprintln(v, "input")
	}
	return nil
}

func updateResultView(g *gocui.Gui, results []PingResult) {
	msg := ""
	for _, v := range results {
		sum := float64(0)
		for _, ttl := range v.ttls {
			sum += float64(ttl.Microseconds())
		}

		ip := v.IP.String()
		rate := float64(v.ReceivedCount) / (float64(v.ReceivedCount) + float64(v.ErrorCount)) * 100
		avg_ttl := float64(sum) / float64(len(v.ttls)) / 1000
		msg += fmt.Sprintf("%s: avg_ttl=%gms, rate=%.2f, received=%d, error=%d\n", ip, avg_ttl, rate, v.ReceivedCount, v.ErrorCount)
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
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func inputedIp(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func quitInputIpView(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func setInputIpView(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("inputIp", maxX/2-10, maxY/2-1, maxX/2+9, maxY/2+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Please Input IP"
	}

	// フォーカスを変更
	if _, err := g.SetCurrentView("inputIp"); err != nil {
		log.Panicln(err)
		return err
	}

	return nil
}

func initKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.SetKeybinding("inputIp", gocui.KeyEnter, gocui.ModNone, inputedIp); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.SetKeybinding("inputIp", 'q', gocui.ModNone, quitInputIpView); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
		return err
	}

	if err := g.SetKeybinding("", 'n', gocui.ModNone, setInputIpView); err != nil {
		log.Panicln(err)
		return err
	}
	return nil
}

func pingIpv4(ip net.IP, count int) PingResult {
	result := PingResult{IP: ip}

	// listen
	c, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		fmt.Println(err)
	}
	defer c.Close()

	// タイムアウトのための時間
	t := time.Now()
	t = t.Add(time.Duration(1) * time.Second)
	err = c.SetReadDeadline(t)
	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < count; i++ {
		// send ping
		start := time.Now()
		id := rand.Intn(65535)
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID:   id,
				Data: []byte("HELLO-R-U-THERE"),
			},
		}
		wb, err := wm.Marshal(nil)
		if err != nil {
			fmt.Println(err)
		}
		if _, err := c.WriteTo(wb, &net.UDPAddr{IP: ip}); err != nil {
			fmt.Println(err)
		}

		// receive
		rb := make([]byte, 1500)
		//n, peer, err := c.ReadFrom(rb)
		n, _, err := c.ReadFrom(rb)
		if err != nil {
			fmt.Println(err)
			result.ErrorCount++
		}
		duration := time.Since(start)

		rm, err := icmp.ParseMessage(IPV4_ICMP, rb[:n])
		if err != nil {
			fmt.Println(err)
			result.ErrorCount++
		}

		if rm.Type == ipv4.ICMPTypeEchoReply && rm.Body.(*icmp.Echo).ID == id {
			result.ReceivedCount++
			result.ttls = append(result.ttls, duration)
		}
	}

	return result
}
