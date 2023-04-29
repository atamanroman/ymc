package tui

import (
	"fmt"
	"github.com/atamanroman/ymc/src/internal/logging"
	"github.com/atamanroman/ymc/src/musiccast"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"sort"
	"strings"
)

var App *tview.Application
var log = logging.Instance
var devices *tview.List

func init() {
	App = createUi()
}

func createUi() *tview.Application {
	status := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("Status:\nfoo")
	status.SetBackgroundColor(tcell.ColorDefault)

	devices = tview.NewList()
	devices.SetTitle("  Speakers  ")
	//devices.SetBorder(true)
	//devices.SetBorderPadding(1, 1, 2, 2)
	devices.SetBackgroundColor(tcell.ColorDefault)
	devices.SetSelectedBackgroundColor(tcell.ColorHotPink)
	devices.SetSecondaryTextColor(tcell.ColorGrey)

	flex := tview.NewFlex().
		AddItem(devices, 20, 0, true).
		AddItem(status, 0, 30, false)
	flex.SetTitle("  ymc  ")
	flex.SetBorder(true)
	flex.SetBorderColor(tcell.ColorBlack)
	flex.SetBorderPadding(1, 1, 2, 2)
	flex.SetTitleColor(tcell.ColorHotPink)
	flex.SetBackgroundColor(tcell.ColorDefault)
	return tview.NewApplication().SetRoot(flex, true)
}
func UpdateUi(speakers map[string]*musiccast.Speaker) {

	sorted := make([]*musiccast.Speaker, 0)
	for _, spkr := range speakers {
		sorted = append(sorted, spkr)
	}
	sort.Slice(sorted, func(a int, b int) bool {
		return sorted[a].FriendlyName > sorted[b].FriendlyName
	})

	App.QueueUpdateDraw(func() {
		for i, spkr := range sorted {
			count := devices.GetItemCount()
			if i < count {
				friendlyName, _ := devices.GetItemText(i)
				if withoutColor(friendlyName) == withoutColor(spkr.FriendlyName) {
					devices.SetItemText(i, coloredFriendlyName(spkr), statusString(spkr))
				} else {
					devices.InsertItem(i, coloredFriendlyName(spkr), statusString(spkr), 0, nil)
				}
			} else {
				// new item
				devices.AddItem(coloredFriendlyName(spkr), statusString(spkr), 0, nil)
			}
		}
	})
}

func statusString(speaker *musiccast.Speaker) string {
	if speaker.Power == musiccast.Standby {
		return "  Standby"
	}

	var bars string
	if speaker.Volume == 0 {
		bars = "⨯"
	} else {
		volPercent := float32(speaker.Volume) / float32(speaker.MaxVolume)
		numBars := int(volPercent * 10 / 2)
		switch numBars {
		case 0:
			bars = "▁"
		case 1:
			bars = "▁▃"
		case 2:
			bars = "▁▃▅"
		case 3:
			bars = "▁▃▅▇"
		case 4:
			bars = "▁▃▅▇"
		case 5:
			bars = "▇▇▇▇"
		default:
			bars = ""
		}
	}
	return fmt.Sprintf("  ⏵⏸ Spotify %s", bars)
}

func coloredFriendlyName(speaker *musiccast.Speaker) string {
	if speaker.Power == musiccast.Standby {
		return "[darkgray]" + speaker.FriendlyName + "[white]"
	}
	return "[green]" + speaker.FriendlyName + "[white]"

}

func withoutColor(label string) string {
	if strings.Count(label, "[") == 2 && strings.Count(label, "]") == 2 {
		start := strings.IndexByte(label, ']') + 1
		end := strings.LastIndexByte(label, '[')
		return label[start:end]
	}
	return label
}
