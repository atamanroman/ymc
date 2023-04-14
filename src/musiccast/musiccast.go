package musiccast

import (
	"encoding/json"
	"fmt"
	"github.com/atamanroman/musiccast/src/ssdp"
	"io"
	"net/http"
	"time"
)

const MusicCastModel = "MusicCast"
const YamahaManufacturer = "Yamaha Corporation"

type Power string

const (
	Standby Power = "standby"
	On      Power = "on"
)

type MusicCastDevice struct {
	ID                 string
	Power              Power
	BaseUrl            string
	ControlUrl         string
	ExtendedControlUrl string
	FriendlyName       string
	DeviceType         string
	Volume             int8
	MaxVolume          int8
}

func (m MusicCastDevice) String() string {
	str, err := json.Marshal(m)
	if err != nil {
		return "Failed to marshal: " + err.Error()
	}
	return string(str)
}

func StartScan() chan MusicCastDevice {
	//go func() {
	//	for {
	//		ssdp.SendDiscover(l)
	//		time.Sleep(time.Second * 10)
	//	}
	//}()

	ch := make(chan MusicCastDevice)
	ssdpCh := make(chan *ssdp.Service)
	go func() {
		for {
			listen(ssdpCh, ch)
		}
	}()
	go func() {
		for {
			err := ssdp.Search(ssdp.All, 10, "0.0.0.0:0", ssdpCh)
			if err != nil {
				panic(fmt.Errorf("MusicCast discovery failed: %w", err))
			}
		}
	}()
	return ch
}

func listen(inCh chan *ssdp.Service, ch chan MusicCastDevice) {
	println("Listen for SSDP services")
	for {
		select {
		case service := <-inCh:
			fmt.Println("Found SSDP Service: %s", service)

			ssdpService, _ := ssdp.FetchSSDP(service)
			if isYamahaMusicCast(ssdpService) {
				var dev = MusicCastDevice{ssdpService.Device.UDN, Standby, ssdpService.XDevice.UrlBase, "?", "?", ssdpService.Device.FriendlyName, ssdpService.Device.ModelName, 0, 100}
				err := updateStatus(&dev)
				if err != nil {
					fmt.Println("Failed to get status for device:", dev.FriendlyName, err)
					continue
				}
				fmt.Println("Found MusicCast device:", dev.FriendlyName)
				ch <- dev
			} else {
				fmt.Println("Ignore non-MusicCast device:", ssdpService.Device.ModelName)
			}
		case <-time.After(10 * time.Second):
			fmt.Printf("Nothing found for 10s, continue.")
		}
	}
}

func isYamahaMusicCast(ssdpService ssdp.MediaRenderer) bool {
	return ssdpService != ssdp.MediaRenderer{} &&
		ssdpService.Device.Manufacturer == YamahaManufacturer &&
		ssdpService.Device.ModelDescription == MusicCastModel &&
		ssdpService.XDevice != ssdp.MediaRenderer{}.XDevice
}

type yxcStatus struct {
	ResponseCode int    `json:"response_code"`
	Power        Power  `json:"power"`
	Sleep        int    `json:"sleep"`
	Volume       int8   `json:"volume"`
	Mute         bool   `json:"mute"`
	MaxVolume    int8   `json:"max_volume"`
	Input        string `json:"input"`
	InputText    string `json:"input_text"`
}

func updateStatus(device *MusicCastDevice) error {
	// TODO dynamic zone?
	resp, err := http.Get(device.BaseUrl + "YamahaExtendedControl/v1/main/getStatus")
	if err != nil {
		return err
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	status := yxcStatus{}
	err = json.Unmarshal(all, &status)
	if err != nil {
		return err
	}
	if status.ResponseCode != 0 {
		return fmt.Errorf("getStatus returned %d", status.ResponseCode)
	}
	device.Power = status.Power
	device.Volume = status.Volume
	device.MaxVolume = status.MaxVolume
	return nil
}
