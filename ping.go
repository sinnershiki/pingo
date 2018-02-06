package main

import (
	"fmt"
	"math/rand"
	"net"
	//"flag"
	//"reflect"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func ping_ipv4(ip string, ch chan string) {
	const IPV4_ICMP = 1

	c, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		fmt.Println(err)
	}
	defer c.Close()

	id := rand.Intn(65535)
	fmt.Println(id)
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
	if _, err := c.WriteTo(wb, &net.UDPAddr{IP: net.ParseIP(ip)}); err != nil {
		fmt.Println(err)
	}

	for {
		rb := make([]byte, 1500)
		n, peer, err := c.ReadFrom(rb)
		if err != nil {
			fmt.Println(err)
		}
		rm, err := icmp.ParseMessage(IPV4_ICMP, rb[:n])
		if err != nil {
			fmt.Println(err)
		}

		if rm.Type == ipv4.ICMPTypeEchoReply && rm.Body.(*icmp.Echo).ID == id {
			fmt.Println(rm.Body.(*icmp.Echo).ID)
			ch <- "got reflection from " + peer.String() + "\n"
			break
		}

	}

}

func main() {
	ch := make(chan string)
	go ping_ipv4("160.16.87.149", ch)
	go ping_ipv4("127.0.0.1", ch)
	res1, res2 := <-ch, <-ch
	fmt.Println(res1, res2)
}
