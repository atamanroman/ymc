package main

import (
	"fmt"
	"github.com/atamanroman/musiccast/src/musiccast"
	"sort"
	"time"
)

var Speakers = make(map[string]*musiccast.Speaker)

func main() {
	ch := musiccast.StartScan()
	defer musiccast.Close()
	for {
		select {
		case update := <-ch:
			if Speakers[update.ID] == nil {
				if update.PartialUpdate {
					fmt.Println("Ignore event for unknown MusicCast speaker")
					continue
				}
				fmt.Println("Found new MusicCast speaker", update)
				Speakers[update.ID] = update
			} else {
				fmt.Println("Got MusicCast speaker update", update)
				if update.PartialUpdate {
					update.UpdateValues(Speakers[update.ID])
				} else {
					// full update
					Speakers[update.ID] = update
				}
			}
		default:
			fmt.Println("Nothing found - sleep")
			time.Sleep(500 * time.Millisecond)
		}
		drawUi()
	}
}

func drawUi() {
	println("List of Speakers:\n---")
	items := make([]string, 0)
	for _, spkr := range Speakers {
		items = append(items, spkr.FriendlyName+": "+string(spkr.Power))
	}
	sort.Strings(items)
	for idx, spkr := range items {
		fmt.Println("(", idx, ")", spkr)
	}
}
