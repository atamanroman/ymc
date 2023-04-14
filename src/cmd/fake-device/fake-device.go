package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 0,
	})
	if err != nil {
		panic(err)
	}

	for {
		body := make([]byte, 1024)
		_, addr, err := listener.ReadFromUDP(body)
		if len(body) > 0 {
			// advertise
			fmt.Println("Read", string(body))
			_, err = listener.WriteToUDP([]byte("foo"), addr)
		} else {
			fmt.Println("Read nothing")
			time.Sleep(time.Second)
		}
		if err != nil {
			panic("ssdp:discover failed: " + err.Error())
		}
	}
}
