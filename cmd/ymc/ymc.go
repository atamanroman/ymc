package main

import (
	"github.com/atamanroman/ymc/internal/logging"
	"github.com/atamanroman/ymc/internal/tui"
	"github.com/atamanroman/ymc/musiccast"
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

	go func() {
		for {
			select {
			case command := <-tui.CommandChan:
				speaker := Speakers[command.Id]

				// don't control standby speakers except power them on
				if speaker.Power == musiccast.Standby && command.Action != tui.PowerOn {
					continue
				}

				switch command.Action {
				case tui.PowerOn:
					err := musiccast.SetPower(speaker, musiccast.On)
					if err != nil {
						// TODO
						continue
					}
				case tui.PowerOff:
					err := musiccast.SetPower(speaker, musiccast.Standby)
					if err != nil {
						// TODO
						continue
					}
				case tui.VolumeUp:
					err := musiccast.SetVolume(speaker, musiccast.Up, command.Value.(int))
					if err != nil {
						// TODO
						continue
					}
				case tui.VolumeDown:
					err := musiccast.SetVolume(speaker, musiccast.Down, command.Value.(int))
					if err != nil {
						// TODO
						continue
					}
				case tui.MuteToggle:
					err := musiccast.SetMute(speaker, !*speaker.Mute)
					if err != nil {
						// TODO
						continue
					}
				}
			}
		}
	}()

	if err := tui.App.Run(); err != nil {
		panic(err)
	}
}
