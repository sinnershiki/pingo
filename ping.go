package main

import (
	"fmt"
	"math/rand"
	"net"	
	"golang.org/x/net/ipv4"
	"golang.org/x/net/icmp"
)

func main() {
	// switch runtime.GOOS {
	// case "darwin":
	// case "linux":
	// 	log.Println("you may need to adjust the net.ipv4.ping_group_range kernel state")
	// default:
	// 	log.Println("not supported on", runtime.GOOS)
	// 	return
	// }

	c, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		fmt.Println(err)
	}
	defer c.Close()

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: rand.Intn(65535),
			Data: []byte("HELLO-R-U-THERE"),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		fmt.Println(err)
	}
	if _, err := c.WriteTo(wb, &net.UDPAddr{IP: net.ParseIP("127.0.0.1")}); err != nil {
		fmt.Println(err)
	}

	rb := make([]byte, 1500)
	n, peer, err := c.ReadFrom(rb)
	if err != nil {
		fmt.Println(err)
	}
	rm, err := icmp.ParseMessage(1, rb[:n])
	if err != nil {
		fmt.Println(err)
	}


	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		fmt.Println("got reflection from %v", peer)
	default:
		fmt.Println("got %+v; want echo reply", rm)
	}
}
