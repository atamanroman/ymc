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

type Action string

const (
	PowerOn  Action = "PowerOn"
	PowerOff Action = "PowerOff"
)

type SpeakerCommand struct {
	Id     string
	Action Action
}

var App *tview.Application
var CommandChan = make(chan SpeakerCommand)

var log = logging.Instance
var speakerList *tview.List
var mainLayout *tview.Pages
var knownSpeakers = make([]*musiccast.Speaker, 0)

func init() {
	status := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("Status:\nfoo")
	status.SetBackgroundColor(tcell.ColorDefault)

	speakerList = createSpeakerList()
	columnLayout := createColumnLayout(status)
	helpDialog := createHelpDialog()

	mainLayout = tview.NewPages()
	mainLayout.AddPage("main", columnLayout, true, true).AddPage("help", helpDialog, true, false)
	mainLayout.SetBackgroundColor(tcell.ColorDefault)

	App = tview.NewApplication().SetRoot(mainLayout, true)
	App.SetInputCapture(defaultKeys)
}

func UpdateUi(updated map[string]*musiccast.Speaker) {
	sorted := make([]*musiccast.Speaker, 0)
	for _, spkr := range updated {
		sorted = append(sorted, spkr)
	}
	sort.Slice(sorted, func(a int, b int) bool {
		return sorted[a].FriendlyName > sorted[b].FriendlyName
	})
	knownSpeakers = sorted

	App.QueueUpdateDraw(func() {
		for i, spkr := range sorted {
			count := speakerList.GetItemCount()
			if i < count {
				friendlyName, _ := speakerList.GetItemText(i)
				if withoutColor(friendlyName) == withoutColor(spkr.FriendlyName) {
					speakerList.SetItemText(i, coloredFriendlyName(spkr), statusString(spkr))
				} else {
					speakerList.InsertItem(i, coloredFriendlyName(spkr), statusString(spkr), 0, nil)
				}
			} else {
				// new item
				speakerList.AddItem(coloredFriendlyName(spkr), statusString(spkr), 0, nil)
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
		bars = "0"
	} else if speaker.Mute {
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
	var input string
	if speaker.InputText != "" {
		input = speaker.InputText
	} else {
		input = "???"
	}
	return fmt.Sprintf("  ⏵⏸ %s %s", input, bars)
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

func defaultKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q':
			return tcell.NewEventKey(tcell.KeyESC, ' ', tcell.ModNone)
		case '?':
			mainLayout.ShowPage("help")
			return event
		}
	}
	return event
}
