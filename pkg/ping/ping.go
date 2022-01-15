package ping

import (
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const IPV4_ICMP = 1
const IP_HEADER_LEN = 20

var Pingers []*Pinger

type Pinger struct {
	IP            net.IP
	ID            int
	Timeout       time.Duration
	Data          []byte
	Rtts          []time.Duration
	Count         int
	interval      time.Duration
	ReceivedCount int
	ErrorCount    int
}

func NewPinger(ip net.IP) *Pinger {
	return &Pinger{
		Timeout:       100 * time.Millisecond,
		ID:            os.Getpid(),
		Data:          []byte("HELLO-R-U-THERE"),
		IP:            ip,
		Count:         5,
		interval:      100 * time.Millisecond,
		ReceivedCount: 0,
		ErrorCount:    0,
	}
}

func RemoveByIndex(i int) {
	Pingers = Pingers[:i+copy(Pingers[i:], Pingers[i+1:])]
}

func (p *Pinger) PingV4() error {
	// listen
	c, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		log.Println(err)
	}
	defer c.Close()

	for i := 0; i < p.Count; i++ {
		start := time.Now()

		wb, err := p.sendICMPPacket(c, i)
		if err != nil {
			return err
		}
		if err := p.receiveICMPPacket(c, i, wb); err != nil {
			p.Rtts = append(p.Rtts, -1)
			p.ErrorCount++
		} else {
			duration := time.Since(start)
			p.Rtts = append(p.Rtts, duration)
			p.ReceivedCount++
		}
		time.Sleep(p.interval)
	}

	return nil
}

func (p *Pinger) sendICMPPacket(c *icmp.PacketConn, seq int) ([]byte, error) {
	// タイムアウトのための時間
	if err := c.SetReadDeadline(time.Now().Add(p.Timeout)); err != nil {
		log.Println(err)
	}

	// send ping
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   p.ID,
			Data: p.Data,
			Seq:  seq,
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		return nil, err
	}

	if n, err := c.WriteTo(wb, &net.UDPAddr{IP: p.IP}); err != nil {
		return nil, err
	} else if n != len(wb) {
		return nil, err
	}

	return wb, nil
}

func (p *Pinger) receiveICMPPacket(c *icmp.PacketConn, seq int, wb []byte) error {
	// receive
	rb := make([]byte, 1500)

	// wait for reply
	for {
		n, peer, err := c.ReadFrom(rb)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				log.Printf("Timeout reading from socket for %s[%d]: %s\n", p.IP, seq, err)
			}
			log.Printf("Error reading from socket for %s: %s\n", p.IP, err)
			return err
		}

		n0 := 0
		if len(wb)+IP_HEADER_LEN == n {
			n0 = IP_HEADER_LEN
		}
		m, err := icmp.ParseMessage(IPV4_ICMP, rb[n0:n])
		if err != nil {
			log.Printf("Error ParseMessage from socket for %s: %s\n", peer.String(), err)
			return err
		}

		if m.Type == ipv4.ICMPTypeEchoReply && m.Body.(*icmp.Echo).ID == p.ID {
			return nil
		}
	}
}
