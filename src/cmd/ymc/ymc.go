package main

import (
	"github.com/atamanroman/ymc/src/internal/logging"
	"github.com/atamanroman/ymc/src/internal/tui"
	"github.com/atamanroman/ymc/src/musiccast"
	"time"
)

var log = logging.Instance
var Speakers = make(map[string]*musiccast.Speaker)

func main() {
	defer logging.Close()
	defer musiccast.Close()
	ch := musiccast.StartScan()
	go func() {
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

			tui.UpdateUi(Speakers)
		}
	}()
	if err := tui.App.Run(); err != nil {
		panic(err)
	}
}
