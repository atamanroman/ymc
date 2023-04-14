package main

import (
	"fmt"
	"github.com/atamanroman/musiccast/src/musiccast"
)

var Devices = make(map[string]musiccast.MusicCastDevice)

func main() {
	ch := musiccast.StartScan()
	defer close(ch)
	for {
		device := <-ch
		// TODO can be initial find or device update
		fmt.Println("Found MusicCast device:", device)
		Devices[device.ID] = device
	}
}
