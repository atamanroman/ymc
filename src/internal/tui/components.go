package tui

import (
	"github.com/atamanroman/ymc/src/musiccast"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

const (
	transparent = tcell.ColorDefault
	light       = tcell.ColorLightGray
	dark        = tcell.ColorGray
	black       = tcell.ColorBlack
	good        = tcell.ColorGrey
	bad         = tcell.ColorRed
	accent      = tcell.ColorHotPink
)

func createFrame() *tview.Frame {
	frame := tview.NewFrame(speakerList)
	frame.AddText("Speakers", true, 0, accent)
	style(frame, "ymc")
	return frame
}

func createSpeakerList() *tview.List {
	devices := tview.NewList()
	style(devices, "")
	devices.SetDoneFunc(func() {
		App.Stop()
	})
	devices.SetSelectedFunc(func(index int, friendlyName string, _ string, _ rune) {
		speaker := knownSpeakers[index]
		var action Action
		if speaker.Power == musiccast.On {
			action = PowerOff
		} else {
			action = PowerOn
		}
		CommandChan <- SpeakerCommand{Id: speaker.ID, Action: action}
	})
	devices.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		index := devices.GetCurrentItem()
		speakerId := knownSpeakers[index].ID
		isShift := event.Modifiers()&tcell.ModShift > 0

		switch event.Key() {
		case tcell.KeyLeft:
			value := 5
			if isShift {
				value = 1
			}
			CommandChan <- SpeakerCommand{speakerId, VolumeDown, value}
			return nil
		case tcell.KeyRight:
			value := 5
			if isShift {
				value = 1
			}
			CommandChan <- SpeakerCommand{speakerId, VolumeUp, value}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'm':
				CommandChan <- SpeakerCommand{speakerId, MuteToggle, nil}
				return nil
			}
		}
		return event
	})

	return devices
}

func createHelpDialog() *tview.Flex {
	// 19 chars wide
	help := strings.TrimSpace(`
RET     Turn on/off
→        Volume up*
←      Volume down*
m       Toggle mute

?         Show help
q              Quit

*Shift: small steps
`)
	helpText := tview.NewTextView().SetText(help)

	style(helpText, "Help (?)")
	helpText.SetDoneFunc(func(_ tcell.Key) {
		mainLayout.SwitchToPage("main")
	})

	// center the text
	helpFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(helpText, 13, 1, true).
			AddItem(nil, 0, 1, false), 23, 1, true).
		AddItem(nil, 0, 1, false)
	return helpFlex
}

func style(layout any, title string) {
	switch x := layout.(type) {
	case *tview.Frame:
		if title != "" {
			x.SetTitle("  " + title + "  ")
		}
		x.SetBorder(true)
		x.SetBorderColor(black)
		x.SetTitleColor(accent)
		x.SetBackgroundColor(transparent)
	case *tview.TextView:
		if title != "" {
			x.SetTitle("  " + title + "  ")
		}
		x.SetBorder(true)
		x.SetBorderColor(light)
		x.SetTitleColor(accent)
		x.SetTextColor(dark)
		x.SetBackgroundColor(transparent)
		x.SetBorderPadding(1, 1, 1, 1)
	case *tview.List:
		if title != "" {
			x.SetTitle("  " + title + "  ")
		}
		x.SetBorder(false)
		x.SetBorderColor(light)
		x.SetBackgroundColor(transparent)
		x.SetSelectedBackgroundColor(accent)
		x.SetSecondaryTextColor(light)
		x.SetMainTextColor(dark)
	}
}
