package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const IPV4_ICMP = 1
const IP_HEADER_LEN = 20

var results []PingResult

type PingResult struct {
	IP            net.IP
	ErrorCount    int
	ReceivedCount int
	ttls          []time.Duration
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Println(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := initKeybindings(g); err != nil {
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

func updateResultView(g *gocui.Gui, results []PingResult) error {
	x0, _, x1, _, err := g.ViewPosition("result")
	if err != nil {
		return err
	}
	msg := "[Num] <Dest IP>: Average TTL, Reached ping rate, Unreached ping count\n"
	msg += strings.Repeat("-", x1-x0) + "\n"
	for i, v := range results {
		sum := float64(0)
		for _, ttl := range v.ttls {
			sum += float64(ttl.Microseconds())
		}

		ip := v.IP.String()
		rate := float64(v.ReceivedCount) / (float64(v.ReceivedCount) + float64(v.ErrorCount)) * 100
		avg_ttl := float64(sum) / float64(len(v.ttls)) / 1000
		msg += fmt.Sprintf("[%d] %s: avg_ttl=%gms, rate=%.2f, received=%d, error=%d\n", i+1, ip, avg_ttl, rate, v.ReceivedCount, v.ErrorCount)
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

	result := pingIpv4(ip, 5)
	results = append(results, result)

	if err := updateResultView(g, results); err != nil {
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

	results = remove(results, num-1)

	if err := updateResultView(g, results); err != nil {
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

func initKeybindings(g *gocui.Gui) error {
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

func remove(slice []PingResult, i int) []PingResult {
	return slice[:i+copy(slice[i:], slice[i+1:])]
}

func pingIpv4(ip net.IP, count int) PingResult {
	id := os.Getpid()
	timeout := 100 * time.Millisecond
	sleep := 100 * time.Microsecond

	result := PingResult{IP: ip}

	// listen
	c, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		log.Println(err)
	}
	defer c.Close()

	for i := 0; i < count; i++ {
		// タイムアウトのための時間
		err = c.SetReadDeadline(time.Now().Add(timeout))
		if err != nil {
			log.Println(err)
		}

		// send ping
		start := time.Now()
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID:   id,
				Data: []byte("HELLO-R-U-THERE"),
				Seq:  i + 1,
			},
		}
		wb, err := wm.Marshal(nil)
		if err != nil {
			log.Println(err)
		}
		if n, err := c.WriteTo(wb, &net.UDPAddr{IP: ip}); err != nil {
			log.Println(err)
		} else if n != len(wb) {
			log.Printf("got %v; want %v\n", n, len(wb))
		}

		// receive
		rb := make([]byte, 1500)
		// wait for reply
		for {
			n, peer, err := c.ReadFrom(rb)
			if err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
					log.Printf("Timeout reading from socket for %s[%d]: %s\n", ip, i, err)
					result.ErrorCount++
					break
				}
				log.Printf("Error reading from socket for %s: %s\n", ip, err)
				break
			}

			n0 := 0
			if len(wb)+IP_HEADER_LEN == n {
				n0 = IP_HEADER_LEN
			}
			m, err := icmp.ParseMessage(IPV4_ICMP, rb[n0:n])
			if err != nil {
				log.Printf("Error ParseMessage from socket for %s: %s\n", peer.String(), err)
				result.ErrorCount++
				break
			}

			log.Println(m)
			if m.Type == ipv4.ICMPTypeEchoReply && m.Body.(*icmp.Echo).ID == id {
				duration := time.Since(start)
				result.ReceivedCount++
				result.ttls = append(result.ttls, duration)
				time.Sleep(sleep)
				break
			}
		}

	}

	return result
}
