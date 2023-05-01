package musiccast

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var httpClient = &http.Client{}

type ErrorCode interface {
	ErrorCode() int
}

type ApiResponse struct {
	ResponseCode int `json:"response_code"`
}

func (r ApiResponse) ErrorCode() int {
	return r.ResponseCode
}

type StatusResponse struct {
	ApiResponse
	Power     Power  `json:"power"`
	Sleep     int    `json:"sleep"`
	Volume    int8   `json:"volume"`
	Mute      bool   `json:"mute"`
	MaxVolume int8   `json:"max_volume"`
	Input     string `json:"input"`
	InputText string `json:"input_text"`
}

func (r StatusResponse) ErrorCode() int {
	return r.ResponseCode
}

// TODO events time out after 10min!

// fetch the current speaker status from the speaker and subscribe to MusicCast events if port > 0
func updateStatus(speaker *Speaker, appPort int) error {
	// TODO dynamic zone?
	status, err := GetStatus(speaker, appPort)
	if err != nil {
		return err
	}
	speaker.Power = status.Power
	speaker.Volume = &status.Volume
	speaker.MaxVolume = status.MaxVolume
	speaker.InputText = status.InputText
	speaker.Input = status.Input
	speaker.Mute = &status.Mute
	return nil
}

func GetStatus(speaker *Speaker, appPort int) (*StatusResponse, error) {
	request, _ := http.NewRequest("GET", speaker.BaseUrl+"YamahaExtendedControl/v1/main/getStatus", nil)
	subscribeEvents(appPort, request)
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	target := StatusResponse{}
	err = unmarshalApiResponse(resp, &target)
	if err != nil {
		return nil, err
	}
	return &target, nil
}

type DeviceInfoResponse struct {
	ApiResponse
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

func (r DeviceInfoResponse) ErrorCode() int {
	return r.ResponseCode
}

func updateDeviceInfo(speaker *Speaker, appPort int) error {
	deviceInfo, err := GetDeviceInfo(speaker, appPort)
	if err != nil {
		return err
	}

	// TODO more?
	speaker.ID = deviceInfo.DeviceId
	return nil
}

func GetDeviceInfo(speaker *Speaker, appPort int) (*DeviceInfoResponse, error) {
	request, _ := http.NewRequest(http.MethodGet, speaker.BaseUrl+"YamahaExtendedControl/v1/system/getDeviceInfo", nil)
	subscribeEvents(appPort, request)
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	target := DeviceInfoResponse{}
	err = unmarshalApiResponse(resp, &target)
	if err != nil {
		return nil, err
	}
	return &target, nil
}

type RangeStep struct {
	Id   string `json:"id"`
	Min  int    `json:"min"`
	Max  int    `json:"max"`
	Step int    `json:"step"`
}

type GetFeaturesResponse struct {
	ApiResponse
	System struct {
		FuncList  []string `json:"func_list"`
		ZoneNum   int      `json:"zone_num"`
		InputList []struct {
			Id                 string `json:"id"`
			DistributionEnable bool   `json:"distribution_enable"`
			RenameEnable       bool   `json:"rename_enable"`
			AccountEnable      bool   `json:"account_enable"`
			PlayInfoType       string `json:"play_info_type"`
		} `json:"input_list"`
		RangeStep []RangeStep `json:"range_step"`
		Bluetooth struct {
			UpdateCancelable      bool `json:"update_cancelable"`
			TxConnectivityTypeMax int  `json:"tx_connectivity_type_max"`
		} `json:"bluetooth"`
	} `json:"system"`
	Zone []struct {
		Id                 string      `json:"id"`
		FuncList           []string    `json:"func_list"`
		InputList          []string    `json:"input_list"`
		SoundProgramList   []string    `json:"sound_program_list"`
		EqualizerModeList  []string    `json:"equalizer_mode_list"`
		LinkControlList    []string    `json:"link_control_list"`
		LinkAudioDelayList []string    `json:"link_audio_delay_list"`
		RangeStep          []RangeStep `json:"range_step"`
		CcsSupported       []string    `json:"ccs_supported"`
	} `json:"zone"`
	Netusb struct {
		FuncList []string `json:"func_list"`
		Preset   struct {
			Num int `json:"num"`
		} `json:"preset"`
		RecentInfo struct {
			Num int `json:"num"`
		} `json:"recent_info"`
		PlayQueue struct {
			Size int `json:"size"`
		} `json:"play_queue"`
		McPlaylist struct {
			Size int `json:"size"`
			Num  int `json:"num"`
		} `json:"mc_playlist"`
		NetRadioType string `json:"net_radio_type"`
		Tidal        struct {
			Mode string `json:"mode"`
		} `json:"tidal"`
		Qobuz struct {
			LoginType string `json:"login_type"`
		} `json:"qobuz"`
	} `json:"netusb"`
	Distribution struct {
		Version          int      `json:"version"`
		CompatibleClient []int    `json:"compatible_client"`
		ClientMax        int      `json:"client_max"`
		ServerZoneList   []string `json:"server_zone_list"`
		McSurround       struct {
			Version    int      `json:"version"`
			FuncList   []string `json:"func_list"`
			MasterRole struct {
				SurroundPair  bool `json:"surround_pair"`
				StereoPair    bool `json:"stereo_pair"`
				SubwooferPair bool `json:"subwoofer_pair"`
			} `json:"master_role"`
			SlaveRole struct {
				SurroundPairLOrR bool `json:"surround_pair_l_or_r"`
				SurroundPairLr   bool `json:"surround_pair_lr"`
				SubwooferPair    bool `json:"subwoofer_pair"`
			} `json:"slave_role"`
		} `json:"mc_surround"`
	} `json:"distribution"`
	Clock struct {
		FuncList         []string    `json:"func_list"`
		RangeStep        []RangeStep `json:"range_step"`
		AlarmFadeTypeNum int         `json:"alarm_fade_type_num"`
		AlarmModeList    []string    `json:"alarm_mode_list"`
		AlarmInputList   []string    `json:"alarm_input_list"`
		AlarmPresetList  []string    `json:"alarm_preset_list"`
	} `json:"clock"`
	Ccs struct {
		Supported bool `json:"supported"`
	} `json:"ccs"`
}

func (r GetFeaturesResponse) ErrorCode() int {
	return r.ResponseCode
}

func GetFeatures(speaker *Speaker) (*GetFeaturesResponse, error) {
	request, _ := http.NewRequest(http.MethodGet, speaker.BaseUrl+"YamahaExtendedControl/v1/system/getFeatures", nil)
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	target := GetFeaturesResponse{}
	err = unmarshalApiResponse(resp, &target)
	if err != nil {
		return nil, err
	}
	return &target, nil
}

func SetPower(speaker *Speaker, power Power) error {
	request, _ := http.NewRequest(http.MethodGet, speaker.BaseUrl+"YamahaExtendedControl/v1/main/setPower?power="+strings.ToLower(string(power)), nil)
	resp, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	target := ApiResponse{}
	err = unmarshalApiResponse(resp, &target)
	if err != nil {
		return err
	}
	return nil
}

func SetVolume(speaker *Speaker, direction Volume, step int) error {
	url := speaker.BaseUrl + "YamahaExtendedControl/v1/main/setVolume?volume=" + strings.ToLower(string(direction))
	if step > 1 {
		url += "&step=" + strconv.Itoa(step)
	}
	request, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	target := ApiResponse{}
	err = unmarshalApiResponse(resp, &target)
	if err != nil {
		return err
	}
	return nil
}

func SetMute(speaker *Speaker, mute bool) error {
	request, _ := http.NewRequest(http.MethodGet, speaker.BaseUrl+"YamahaExtendedControl/v1/main/setMute?enable="+strconv.FormatBool(mute), nil)
	resp, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	target := ApiResponse{}
	err = unmarshalApiResponse(resp, &target)
	if err != nil {
		return err
	}
	return nil
}

type GetPlayInfoResponse struct {
	ApiResponse
	Input         string `json:"input"`
	Playback      string `json:"playback"`
	Repeat        string `json:"repeat"`
	Shuffle       string `json:"shuffle"`
	PlayTime      int    `json:"play_time"`
	TotalTime     int    `json:"total_time"`
	Artist        string `json:"artist"`
	Album         string `json:"album"`
	Track         string `json:"track"`
	AlbumartUrl   string `json:"albumart_url"`
	AlbumartId    int    `json:" albumart_id"`
	UsbDevicetype string `json:"usb_devicetype"`
	Attribute     int    `json:"attribute"`
}

func (o GetPlayInfoResponse) ErrorCode() int {
	return o.ResponseCode
}

func GetPlayInfo(speaker *Speaker) (*GetPlayInfoResponse, error) {
	request, _ := http.NewRequest(http.MethodGet, speaker.BaseUrl+"YamahaExtendedControl/v1/netusb/getPlayInfo", nil)
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	target := GetPlayInfoResponse{}
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

func unmarshalApiResponse(resp *http.Response, target ErrorCode) error {
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
