package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var ipNum = 0

func ping_ipv4(ip string, ch chan string) {
	const IPV4_ICMP = 1

	c, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		fmt.Println(err)
	}
	defer c.Close()

	t := time.Now()
	t = t.Add(time.Duration(1) * time.Second)

	err = c.SetReadDeadline(t)
	if err != nil {
		fmt.Println(err)
	}

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
	if _, err := c.WriteTo(wb, &net.UDPAddr{IP: net.ParseIP(ip)}); err != nil {
		fmt.Println(err)
	}

	for {
		rb := make([]byte, 1500)
		n, peer, err := c.ReadFrom(rb)
		if err != nil {
			//fmt.Println(err)
			fmt.Println("timeout")
			ipNum--
			break
		}
		rm, err := icmp.ParseMessage(IPV4_ICMP, rb[:n])
		if err != nil {
			fmt.Println(err)
		}

		if rm.Type == ipv4.ICMPTypeEchoReply && rm.Body.(*icmp.Echo).ID == id {
			ch <- "got reflection from " + peer.String() + "\n"
			ipNum--
			break
		}
	}
	if ipNum == 0 {
		close(ch)
	}
}

func main() {
	flag.Parse()
	ips := flag.Args()
	fmt.Println(ips)
	ipNum = len(ips)

	ch := make(chan string)

	for _, ip := range ips {
		go ping_ipv4(ip, ch)
	}

	for elem := range ch {
		fmt.Println(elem)
	}
}
