package ssdp

import (
	"encoding/json"
	"encoding/xml"
	"github.com/atamanroman/musiccast/src/ssdp/multicast"
	"github.com/atamanroman/musiccast/src/ssdp/ssdplog"
	"io"
	"log"
	"net"
	"net/http"
)

func init() {
	multicast.InterfacesProvider = func() []net.Interface {
		return Interfaces
	}
	ssdplog.LoggerProvider = func() *log.Logger {
		return Logger
	}
}

// Interfaces specify target interfaces to multicast.  If no interfaces are
// specified, all interfaces will be used.
var Interfaces []net.Interface

// Logger is default logger for SSDP module.
var Logger *log.Logger

const UpnpMediaRenderer = "urn:schemas-upnp-org:device:MediaRenderer:1"

type MediaRenderer struct {
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

func (s MediaRenderer) String() string {
	str, err := json.Marshal(s)
	if err != nil {
		return "Failed to marshal: " + err.Error()
	}
	return string(str)
}

// SetMulticastRecvAddrIPv4 updates multicast address where to receive packets.
// This never fail now.
func SetMulticastRecvAddrIPv4(addr string) error {
	return multicast.SetRecvAddrIPv4(addr)
}

func FetchSSDP(device *Service) (MediaRenderer, error) {
	println("Fetch SSDP info for %s from %s", device.USN, device.Location)
	resp, err := http.Get(device.Location)
	if err != nil {
		return MediaRenderer{}, err
	}
	all, err := io.ReadAll(resp.Body)
	//fmt.Println(string(all))
	defer resp.Body.Close()
	if err != nil {
		return MediaRenderer{}, err
	}

	var ssdpService = MediaRenderer{}
	err = xml.Unmarshal(all, &ssdpService)
	if err != nil {
		return MediaRenderer{}, err
	}
	return ssdpService, nil
}

// SetMulticastSendAddrIPv4 updates a UDP address to send multicast packets.
// This never fail now.
func SetMulticastSendAddrIPv4(addr string) error {
	return multicast.SetSendAddrIPv4(addr)
}
