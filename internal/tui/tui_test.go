package tui

import (
	"github.com/atamanroman/ymc/internal/testhelper"
	"github.com/atamanroman/ymc/musiccast"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestStatusString(t *testing.T) {
	speaker := musiccast.Speaker{}

	speaker.Volume = testhelper.Ptr(int8(30))
	speaker.MaxVolume = 100

	assert.Equal(t, "⏵⏸ ??? ◢ 30%", trimmedStatus(speaker))

	speaker.Mute = testhelper.Ptr(true)
	assert.Equal(t, "⏵⏸ ??? ◢ M", trimmedStatus(speaker))

	speaker.Volume = testhelper.Ptr(int8(0))
	assert.Equal(t, "⏵⏸ ??? ◢ M", trimmedStatus(speaker))

	speaker.Mute = nil
	assert.Equal(t, "⏵⏸ ??? ◢ 0%", trimmedStatus(speaker))

	speaker.InputText = "Net Radio"
	assert.Equal(t, "⏵⏸ Net Radio ◢ 0%", trimmedStatus(speaker))
}

func trimmedStatus(speaker musiccast.Speaker) string {
	return strings.TrimSpace(statusString(&speaker))
}
