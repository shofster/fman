package main

/*

  File:    commandBar.go
  Author:  Bob Shofner

  Copyright (c) 2020. See MIT-LICENSE

  The above copyright file and this permission notice shall be included in all
  copies or substantial portions of the Software.

    Build a commandBar for the application.

*/

import (
	"fman/control"
	"fman/element"
	"fman/sys"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
)

func newCommandBar(panelA *control.Panel, panelB *control.Panel) *fyne.Container {
	system := sys.GetSystem()

	// bottom controls for each panel:
	// change preferencesbutton and the show hidden files selection
	settings := widget.NewButton("", func() {
		control.ShowAndEditPrefs(system, panelA, panelB)
	})
	settings.SetIcon(theme.SettingsIcon())

	// command line Button
	cline := widget.NewButton("    Command Line    ", func() {
		cmd := system.Settings.GetShellCommand()
		if cmd == nil {
			sys.Toast("Select a Place", sys.WarnToast)
		} else {
			err := cmd.Start()
			if err != nil {
				system.Cline.SetIcon(theme.WarningIcon())
				s := fmt.Sprintf("Command Error: %s\n", err)
				sys.Toast(s, sys.WarnToast)
				log.Printf(s)
				return
			}
			system.Cline.SetIcon(theme.ComputerIcon())
			return
		}

		system.Cline.SetIcon(theme.WarningIcon())
	})
	system.Cline = cline
	system.Cline.SetIcon(theme.ComputerIcon())

	// label and infinite progress widget
	active := widget.NewLabel("  Busy: ")
	system.BusyIndicator = widget.NewProgressBarInfinite()
	system.BusyIndicator.Stop()

	// font size
	small := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
		element.TextMinus()
	})
	big := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
		element.TextPlus()
	})
	size := widget.NewLabel("font")
	font := container.NewHBox(small, size, big)

	//
	return container.NewHBox(widget.NewLabel(" Preferences: "), settings,
		widget.NewLabel("   "), sys.GetHidden(system.Settings),
		sys.GetDateTime(system.Settings),
		sys.GetDescending(system.Settings),
		widget.NewLabel(" "), font, widget.NewLabel("   "),
		cline, active, system.BusyIndicator)
	//
}
