package ssdp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/atamanroman/ymc/internal/logging"
	multicast2 "github.com/atamanroman/ymc/internal/ssdp/multicast"
	"net"
	"net/http"
	"time"
)

const (
	// All is a search type to search all services and devices.
	All = "ssdp:all"

	// RootDevice is a search type to search UPnP root devices.
	RootDevice        = "upnp:rootdevice"
	UpnpMediaRenderer = "urn:schemas-upnp-org:device:MediaRenderer:1"
)

// Search searches services by SSDP.
func Search(searchType string, waitSec int, ch chan<- *Service) error {
	// dial multicast UDP packet.
	conn, err := multicast2.Listen(&multicast2.AddrResolver{Addr: "0.0.0.0:0"})
	if err != nil {
		return err
	}
	defer conn.Close()
	logging.Instance.Debugf("search on %s", conn.LocalAddr().String())

	// send request.
	addr, err := multicast2.SendAddr()
	if err != nil {
		return err
	}
	msg, err := buildSearch(addr, searchType, waitSec)
	if err != nil {
		return err
	}
	if _, err := conn.WriteTo(multicast2.BytesDataProvider(msg), addr); err != nil {
		return err
	}

	h := func(a net.Addr, d []byte) error {
		srv, err := parseService(a, d)
		if err != nil {
			logging.Instance.Debugf("invalid search response from %s: %s", a.String(), err)
			return nil
		}
		logging.Instance.Debugf("search response from %s: %s", a.String(), srv.USN)
		ch <- srv
		return nil
	}
	d := time.Second * time.Duration(waitSec)
	if err := conn.ReadPackets(d, h); err != nil {
		return err
	}

	return err
}

func buildSearch(raddr net.Addr, searchType string, waitSec int) ([]byte, error) {
	b := new(bytes.Buffer)
	// FIXME: error should be checked.
	b.WriteString("M-SEARCH * HTTP/1.1\r\n")
	fmt.Fprintf(b, "HOST: %s\r\n", raddr.String())
	fmt.Fprintf(b, "MAN: %q\r\n", "ssdp:discover")
	fmt.Fprintf(b, "MX: %d\r\n", waitSec)
	fmt.Fprintf(b, "ST: %s\r\n", searchType)
	b.WriteString("\r\n")
	return b.Bytes(), nil
}

var (
	errWithoutHTTPPrefix = errors.New("without HTTP prefix")
)

var endOfHeader = []byte{'\r', '\n', '\r', '\n'}

func parseService(addr net.Addr, data []byte) (*Service, error) {
	if !bytes.HasPrefix(data, []byte("HTTP")) {
		return nil, errWithoutHTTPPrefix
	}
	// Complement newlines on tail of header for buggy SSDP responses.
	if !bytes.HasSuffix(data, endOfHeader) {
		// why we should't use append() for this purpose:
		// https://play.golang.org/p/IM1pONW9lqm
		data = bytes.Join([][]byte{data, endOfHeader}, nil)
	}
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(data)), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return &Service{
		Type:      resp.Header.Get("ST"),
		USN:       resp.Header.Get("USN"),
		Location:  resp.Header.Get("LOCATION"),
		Server:    resp.Header.Get("SERVER"),
		rawHeader: resp.Header,
	}, nil
}
