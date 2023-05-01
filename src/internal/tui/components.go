package tui

import (
	"github.com/atamanroman/ymc/src/musiccast"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func createColumnLayout(status *tview.TextView) *tview.Flex {
	columnLayout := tview.NewFlex().
		AddItem(speakerList, 20, 0, true).
		AddItem(status, 0, 30, false)
	columnLayout.SetTitle("  ymc  ")
	columnLayout.SetBorder(true)
	columnLayout.SetBorderColor(tcell.ColorBlack)
	columnLayout.SetBorderPadding(1, 1, 2, 2)
	columnLayout.SetTitleColor(tcell.ColorHotPink)
	columnLayout.SetBackgroundColor(tcell.ColorDefault)
	return columnLayout
}

func createSpeakerList() *tview.List {
	devices := tview.NewList()
	devices.SetTitle("  Speakers  ")
	//devices.SetBorder(true)
	//devices.SetBorderPadding(1, 1, 2, 2)
	devices.SetBackgroundColor(tcell.ColorDefault)
	devices.SetSelectedBackgroundColor(tcell.ColorHotPink)
	devices.SetSecondaryTextColor(tcell.ColorGrey)
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
		switch event.Key() {
		case tcell.KeyLeft:
			CommandChan <- SpeakerCommand{speakerId, VolumeDown}
			return nil
		case tcell.KeyRight:
			CommandChan <- SpeakerCommand{speakerId, VolumeUp}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'm':
				CommandChan <- SpeakerCommand{speakerId, MuteToggle}
				return nil
			}
		}
		return event
	})

	return devices
}

func createHelpDialog() *tview.Flex {
	helpText := tview.NewTextView().SetText(
		"RET\tTurn on/off\n" +
			"→\t\tVolume up\n" +
			"←\t\tVolume down\n" +
			"m\t\tToggle mute\n" +
			"?\t\tShow help\n" +
			"q\t\tQuit\n")
	helpText.SetTitle("  ymc Help (?)  ")
	helpText.SetBorder(true)
	helpText.SetBackgroundColor(tcell.ColorDefault)
	helpText.SetDoneFunc(func(_ tcell.Key) {
		mainLayout.SwitchToPage("main")
	})
	helpText.SetBorderPadding(1, 1, 1, 1)

	// center the text
	helpFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(helpText, 9, 1, true).
			AddItem(nil, 0, 1, false), 30, 1, true).
		AddItem(nil, 0, 1, false)
	return helpFlex
}