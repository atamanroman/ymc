package musiccast

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

var httpClient = &http.Client{}

type ApiResponse interface {
	ErrorCode() int
}

type statusResponse struct {
	ResponseCode int    `json:"response_code"`
	Power        Power  `json:"power"`
	Sleep        int    `json:"sleep"`
	Volume       int8   `json:"volume"`
	Mute         bool   `json:"mute"`
	MaxVolume    int8   `json:"max_volume"`
	Input        string `json:"input"`
	InputText    string `json:"input_text"`
}

func (r statusResponse) ErrorCode() int {
	return r.ResponseCode
}

// TODO events time out after 10min!

// fetch the current speaker status from the device and subscribe to MusicCast events if port > 0
func updateStatus(device *Speaker, appPort int) error {
	// TODO dynamic zone?
	status, err := getStatus(device, appPort)
	if err != nil {
		return err
	}
	device.Power = status.Power
	device.Volume = status.Volume
	device.MaxVolume = status.MaxVolume
	return nil
}

func getStatus(device *Speaker, appPort int) (*statusResponse, error) {
	request, _ := http.NewRequest("GET", device.BaseUrl+"YamahaExtendedControl/v1/main/getStatus", nil)
	subscribeEvents(appPort, request)
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	target := statusResponse{}
	err = unmarshalApiResponse(resp, &target)
	if err != nil {
		return nil, err
	}
	return &target, nil
}

type deviceInfoResponse struct {
	ResponseCode        int     `json:"response_code"`
	ModelName           string  `json:"model_name"`
	Destination         string  `json:"destination"`
	DeviceId            string  `json:"device_id"`
	SystemId            string  `json:"system_id"`
	SystemVersion       float64 `json:"system_version"`
	ApiVersion          float64 `json:"api_version"`
	NetmoduleGeneration int     `json:"netmodule_generation"`
	NetmoduleVersion    string  `json:"netmodule_version"`
	NetmoduleChecksum   string  `json:"netmodule_checksum"`
	SerialNumber        string  `json:"serial_number"`
	CategoryCode        int     `json:"category_code"`
	OperationMode       string  `json:"operation_mode"`
	UpdateErrorCode     string  `json:"update_error_code"`
	UpdateDataType      int     `json:"update_data_type"`
}

func (r deviceInfoResponse) ErrorCode() int {
	return r.ResponseCode
}

func updateDeviceInfo(device *Speaker, appPort int) error {
	deviceInfo, err := getDeviceInfo(device, appPort)
	if err != nil {
		return err
	}

	// TODO more?
	device.ID = deviceInfo.DeviceId
	return nil
}

func getDeviceInfo(device *Speaker, appPort int) (*deviceInfoResponse, error) {
	request, _ := http.NewRequest("GET", device.BaseUrl+"YamahaExtendedControl/v1/system/getDeviceInfo", nil)
	subscribeEvents(appPort, request)
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	target := deviceInfoResponse{}
	err = unmarshalApiResponse(resp, &target)
	if err != nil {
		return nil, err
	}
	return &target, nil
}

func subscribeEvents(appPort int, request *http.Request) {
	if appPort > 0 {
		log.Infof("Subscribe to MusicCast events on port=%d", appPort)
		request.Header.Add("X-AppName", "MusicCast/CLI")
		request.Header.Add("X-AppPort", strconv.Itoa(appPort))
	} else {
		log.Info("Skip MusicCast event subscription")
	}
}

func unmarshalApiResponse(resp *http.Response, target ApiResponse) error {
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(all, &target)
	if err != nil {
		return err
	}
	if target.ErrorCode() != 0 {
		return fmt.Errorf("API response returned %d", target.ErrorCode())
	}
	return nil
}
