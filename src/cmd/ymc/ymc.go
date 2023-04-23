package main

import (
	"fmt"
	"github.com/atamanroman/ymc/src/internal/logging"
	"github.com/atamanroman/ymc/src/musiccast"
	"sort"
	"time"
)

var Speakers = make(map[string]*musiccast.Speaker)
var log = logging.Instance

func main() {
	defer logging.Close()
	defer musiccast.Close()
	ch := musiccast.StartScan()
	for {
		select {
		case update := <-ch:
			if Speakers[update.ID] == nil {
				if update.PartialUpdate {
					log.Debug("Ignore event for unknown MusicCast speaker")
					continue
				}
				log.Info("Found new MusicCast speaker", update)
				Speakers[update.ID] = update
			} else {
				log.Debug("Got MusicCast speaker update", update)
				if update.PartialUpdate {
					update.UpdateValues(Speakers[update.ID])
				} else {
					// full update
					Speakers[update.ID] = update
				}
			}
		default:
			log.Debug("Nothing found - sleep")
			time.Sleep(500 * time.Millisecond)
		}
		drawUi()
	}
}

func drawUi() {
	fmt.Println("\033[H\033[2J")
	fmt.Println("List of Speakers:\n---")
	items := make([]string, 0)
	for _, spkr := range Speakers {
		items = append(items, spkr.FriendlyName+": "+string(spkr.Power))
	}
	sort.Strings(items)
	for idx, spkr := range items {
		fmt.Println("(", idx, ")", spkr)
	}
}
