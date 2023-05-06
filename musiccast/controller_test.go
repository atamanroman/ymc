package musiccast

import (
	"github.com/atamanroman/ymc/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateValues(t *testing.T) {
	var speaker = Speaker{
		ID:                 "1",
		Power:              On,
		BaseUrl:            "https://example.com",
		ControlUrl:         "/foo",
		ExtendedControlUrl: "/bar",
		FriendlyName:       "Office",
		DeviceType:         "WX-021",
		Volume:             testhelper.Ptr(int8(30)),
		MaxVolume:          100,
		InputText:          "Digital",
		Input:              "digital",
		Mute:               testhelper.Ptr(false),
		PartialUpdate:      false,
	}
	update := Speaker{
		ID:            "1",
		Power:         Standby,
		Volume:        testhelper.Ptr(int8(70)),
		InputText:     "Net Radio",
		Input:         "netradio",
		Mute:          testhelper.Ptr(true),
		PartialUpdate: true,
	}

	speaker.UpdateValues(&update)

	assert.Equal(t, update.ID, speaker.ID)
	assert.Equal(t, update.Power, speaker.Power)
	assert.Equal(t, update.Volume, speaker.Volume)
	assert.Equal(t, update.InputText, speaker.InputText)
	assert.Equal(t, update.Input, speaker.Input)
	assert.Equal(t, update.Mute, speaker.Mute)
	assert.Equal(t, false, speaker.PartialUpdate)
	assert.Equal(t, "Office", speaker.FriendlyName)
	assert.Equal(t, "WX-021", speaker.DeviceType)
}
