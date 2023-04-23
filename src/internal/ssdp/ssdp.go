package ssdp

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	multicast2 "github.com/atamanroman/musiccast/src/internal/ssdp/multicast"
	"github.com/atamanroman/musiccast/src/internal/ssdp/ssdplog"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
)

func init() {
	multicast2.InterfacesProvider = func() []net.Interface {
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

func (s *MediaRenderer) String() string {
	str, err := json.Marshal(s)
	if err != nil {
		return "Failed to marshal: " + err.Error()
	}
	return string(str)
}

// Service is discovered service.
type Service struct {
	// Type is a property of "ST"
	Type string

	// USN is a property of "USN"
	USN string

	// Location is a property of "LOCATION"
	Location string

	// Server is a property of "SERVER"
	Server string

	rawHeader http.Header
	maxAge    *int
}

func (s *Service) String() string {
	str, err := json.Marshal(s)
	if err != nil {
		return "Failed to marshal: " + err.Error()
	}
	return string(str)
}

var rxMaxAge = regexp.MustCompile(`\bmax-age\s*=\s*(\d+)\b`)

func extractMaxAge(s string, value int) int {
	v := value
	if m := rxMaxAge.FindStringSubmatch(s); m != nil {
		i64, err := strconv.ParseInt(m[1], 10, 32)
		if err == nil {
			v = int(i64)
		}
	}
	return v
}

// MaxAge extracts "max-age" value from "CACHE-CONTROL" property.
func (s *Service) MaxAge() int {
	if s.maxAge == nil {
		s.maxAge = new(int)
		*s.maxAge = extractMaxAge(s.rawHeader.Get("CACHE-CONTROL"), -1)
	}
	return *s.maxAge
}

// Header returns all properties in response of search.
func (s *Service) Header() http.Header {
	return s.rawHeader
}

// SetMulticastRecvAddrIPv4 updates multicast address where to receive packets.
// This never fail now.
func SetMulticastRecvAddrIPv4(addr string) error {
	return multicast2.SetRecvAddrIPv4(addr)
}

func GetMediaRenderer(device *Service) (*MediaRenderer, error) {
	fmt.Printf("Fetch SSDP info for %v from %v", device.USN, device.Location)
	resp, err := http.Get(device.Location)
	if err != nil {
		return nil, err
	}
	all, err := io.ReadAll(resp.Body)
	//fmt.Println(string(all))
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var ssdpService = MediaRenderer{}
	err = xml.Unmarshal(all, &ssdpService)
	if err != nil {
		return nil, err
	}
	return &ssdpService, nil
}

// SetMulticastSendAddrIPv4 updates a UDP address to send multicast packets.
// This never fail now.
func SetMulticastSendAddrIPv4(addr string) error {
	return multicast2.SetSendAddrIPv4(addr)
}
