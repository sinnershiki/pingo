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

type PingResult struct {
	IP            net.IP
	ErrorCount    int
	ReceivedCount int
	TTLs          []time.Duration
}

func PingIpv4(ip net.IP, count int) PingResult {
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
				result.TTLs = append(result.TTLs, duration)
				time.Sleep(sleep)
				break
			}
		}

	}

	return result
}
