package musiccast

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/atamanroman/ymc/internal/logging"
	ssdp2 "github.com/atamanroman/ymc/internal/ssdp"
	"net"
	"strconv"
	"strings"
	"time"
)

var log = logging.Instance

const musicCastModel = "MusicCast"
const musicCastManufacturer = "Yamaha Corporation"

type Power string

const (
	Standby Power = "standby"
	On      Power = "on"
)

type Volume string

const (
	Up   Volume = "Up"
	Down Volume = "Down"
)

type Speaker struct {
	ID                 string
	Power              Power
	BaseUrl            string
	ControlUrl         string
	ExtendedControlUrl string
	FriendlyName       string
	DeviceType         string
	Volume             *int8
	MaxVolume          int8
	InputText          string
	Input              string
	Mute               *bool

	PartialUpdate bool
}

func (o Speaker) String() string {
	return jsonStringer(o)
}

// UpdateValues copies non-empty values onto target
func (o Speaker) UpdateValues(target *Speaker) {
	if o.ID == "" {
		return
	}
	if o.Power != "" {
		target.Power = o.Power
	}

	// TODO 0 is a valid value!
	if o.Volume != nil {
		target.Volume = o.Volume
	}

	if o.InputText != "" {
		target.InputText = o.InputText
	}

	if o.Input != "" {
		target.Input = o.Input
	}

	if o.Mute != nil {
		target.Mute = o.Mute
	}

	// TODO
}

type ZonedStatusEvent struct {
	ID     string      `json:"device_id"`
	Main   StatusEvent `json:"main"`
	Netusb NetusbEvent `json:"netusb"`
}
type StatusEvent struct {
	Power  Power `json:"power"`
	Volume *int8 `json:"volume"`
	Mute   *bool `json:"mute"`
}
type NetusbEvent struct {
	PlayError       *int  `json:"play_error "`
	AccountUpdated  *bool `json:"account_updated"`
	PlayTime        *int  `json:"play_time"`
	PlayInfoUpdated *bool `json:"play_info_updated"`
	ListInfoUpdated *bool `json:"list_info_updated"`
}

func (o ZonedStatusEvent) String() string {
	return jsonStringer(o)
}

func (o StatusEvent) String() string {
	return jsonStringer(o)
}

func (o NetusbEvent) String() string {
	return jsonStringer(o)
}

func jsonStringer(obj any) string {
	str, err := json.Marshal(obj)
	if err != nil {
		log.DPanic("Failed to marshal", err)
		return ""
	}
	return string(str)
}

var ssdpChan = make(chan *ssdp2.Service)
var speakerChan = make(chan *Speaker)
var eventConnection *net.UDPConn
var eventListenerPort int

func init() {
	log.Debug("Init MusicCast client")
	var err error
	eventConnection, err = net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 0})
	if err != nil {
		panic(fmt.Errorf("MusicCast event listener failed: %w", err))
	}
	eventListenerPort, _ = strconv.Atoi(strings.Split(eventConnection.LocalAddr().String(), ":")[1])
}

func StartScan() <-chan *Speaker {
	go func() {
		// multiple times because sometimes speakers seem to be a bit unreliable
		for i := 0; i < 5; i++ {
			log.Info("Send SSDP M-Search ")
			err := ssdp2.Search(ssdp2.UpnpMediaRenderer, 1, ssdpChan)
			if err != nil {
				panic(fmt.Errorf("SSDP M-Search failed: %w", err))
			}
		}
	}()
	go func() {
		mediaRendererToMusicCast(ssdpChan, speakerChan, eventListenerPort)
	}()
	go func() {
		for {
			buf := make([]byte, 65536)
			read, _, err := eventConnection.ReadFromUDP(buf)
			if err != nil {
				panic(fmt.Errorf("listen for multicast event failed: %w", err))
			}
			event := ZonedStatusEvent{}
			err = json.Unmarshal(buf[:read], &event)
			if err != nil || event.ID == "" {
				if event.ID == "" {
					err = errors.New("event has no ID")
				}
				log.Warnf("Discard broken MusicCast event: %s\nPayload:\n---\n%s\n---\n", err, string(buf[:read]))
				continue
			}

			if event.Netusb.PlayTime != nil {
				log.Debug("Discard play_time updates for now")
				continue
			}

			spkr := Speaker{}
			if event.ID != "" {
				spkr.ID = event.ID
			}
			spkr.PartialUpdate = true
			if event.Main.Power != "" {
				spkr.Power = event.Main.Power
			}

			if event.Main.Power != "" {
				spkr.Power = event.Main.Power
			}

			if event.Main.Volume != nil {
				spkr.Volume = event.Main.Volume
			}

			if event.Main.Mute != nil {
				spkr.Mute = event.Main.Mute
			}

			speakerChan <- &spkr
		}
	}()
	return speakerChan
}

func mediaRendererToMusicCast(mediaRendererChan <-chan *ssdp2.Service, speakerChan chan<- *Speaker, musicCastEventPort int) {
	log.Info("Listen for SSDP services")
	for {
		select {
		case service := <-mediaRendererChan:
			log.Infof("Found SSDP Service: %v\n", service)
			mediaRenderer, _ := ssdp2.GetMediaRenderer(service)
			if isYamahaMusicCast(mediaRenderer) {
				var spkr = Speaker{mediaRenderer.Device.UDN, Standby, mediaRenderer.XDevice.UrlBase, "?", "?", mediaRenderer.Device.FriendlyName, mediaRenderer.Device.ModelName, nil, 100, "", "", nil, false}
				err := updateStatus(&spkr, musicCastEventPort)
				if err != nil {
					log.Warn("Failed to get status for device:", spkr.FriendlyName, err)
					continue
				}
				err = updateDeviceInfo(&spkr, musicCastEventPort)
				if err != nil {
					log.Warn("Failed to get deviceInfo for device:", spkr.FriendlyName, err)
					continue
				}
				log.Info("Found MusicCast device:", spkr.FriendlyName)
				speakerChan <- &spkr
			} else {
				log.Debug("Ignore non-MusicCast device:", mediaRenderer.Device.ModelName)
			}
		default:
			//case <-time.After(10 * time.Second):
			log.Debug("No new MediaRenderer found - sleep")
			time.Sleep(1 * time.Second)
		}
	}
}

func isYamahaMusicCast(mediaRenderer *ssdp2.MediaRenderer) bool {
	return mediaRenderer != nil &&
		mediaRenderer.Device.Manufacturer == musicCastManufacturer &&
		mediaRenderer.Device.ModelDescription == musicCastModel &&
		mediaRenderer.XDevice != ssdp2.MediaRenderer{}.XDevice
}

func Close() error {
	// closing eventConnection fails :>
	return nil
}
