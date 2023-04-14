package main

import (
	"encoding/json"
	xml "encoding/xml"
	"fmt"
	"github.com/koron/go-ssdp"
	"io"
	"net"
	"net/http"
)

var Devices = make(map[string]MusicCastDevice)

const MusicCastModel = "MusicCast"
const YamahaManufacturer = "Yamaha Corporation"
const MediaRenderer = "urn:schemas-upnp-org:device:MediaRenderer:1"

type mediaRenderer struct {
	XMLName xml.Name `xml:"root"`
	Device  struct {
		UDN              string `xml:"UDN"`
		FriendlyName     string `xml:"friendlyName"`
		ModelDescription string `xml:"modelDescription"`
		ModelName        string `xml:"modelName"`
		Manufacturer     string `xml:"manufacturer"`
	} `xml:"device"`
	XDevice struct {
		UrlBase    string `xml:"X_URLBase"`
		ControlUrl string `xml:"X_YxcControlURL"`
	} `xml:"X_device"`
}

func (s mediaRenderer) String() string {
	str, err := json.Marshal(s)
	if err != nil {
		return "Failed to marshal: " + err.Error()
	}
	return string(str)
}

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

func StartScan() chan map[string]MusicCastDevice {
	ch := make(chan map[string]MusicCastDevice)
	go func() {
		for {
			scan(ch)
		}
	}()
	return ch
}

func fetchSSDP(device ssdp.Service) (mediaRenderer, error) {
	resp, err := http.Get(device.Location)
	if err != nil {
		return mediaRenderer{}, err
	}
	all, err := io.ReadAll(resp.Body)
	//fmt.Println(string(all))
	defer resp.Body.Close()
	if err != nil {
		return mediaRenderer{}, err
	}

	var ssdpService = mediaRenderer{}
	err = xml.Unmarshal(all, &ssdpService)
	if err != nil {
		return mediaRenderer{}, err
	}
	return ssdpService, nil
}

func scan(ch chan map[string]MusicCastDevice) {

	en0, err := net.InterfaceByName("en0")
	if err != nil {
		panic(err)
	}
	ssdp.Interfaces = []net.Interface{*en0}

	services, err := ssdp.Search(MediaRenderer, 1, "0.0.0.0:0")
	if err != nil {
		panic(err)
	}
	updated := false
	for _, service := range services {
		ssdpService, _ := fetchSSDP(service)
		if isYamahaMusicCast(ssdpService) {
			var dev = MusicCastDevice{ssdpService.Device.UDN, Standby, ssdpService.XDevice.UrlBase, "?", "?", ssdpService.Device.FriendlyName, ssdpService.Device.ModelName, 0, 100}
			err := updateStatus(&dev)
			if err != nil {
				fmt.Println("Failed to get status for device:", dev.FriendlyName, err)
				continue
			}
			Devices[dev.ID] = dev
			updated = true
		} else {
			fmt.Println("Ignore non-MusicCast device:", ssdpService.Device.ModelName)
		}
		if updated {
			ch <- Devices
		}
	}
}

func isYamahaMusicCast(ssdpService mediaRenderer) bool {
	return ssdpService != mediaRenderer{} &&
		ssdpService.Device.Manufacturer == YamahaManufacturer &&
		ssdpService.Device.ModelDescription == MusicCastModel &&
		ssdpService.XDevice != mediaRenderer{}.XDevice
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

func main() {
	ch := StartScan()
	for {
		devices := <-ch
		for _, d := range devices {
			fmt.Println("Found MusicCast device:", d)
		}
	}
}
