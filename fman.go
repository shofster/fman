package main

import (
	"fman/app"
	"fman/control"
	"fman/element"
	"fman/sys"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"log"
	"os"
	"path/filepath"
)

/*

  File:    fman.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/

/*
    A dual pane file manager.

    compile with -ldflags -H=windowsgui (2020 :-))

   fyne bundle gus2.png >> resources.go
   fyne package --release --icon=gus2.png --id=com.scsi.fman

*/

var system *sys.System

func main() {
	// system has global variables
	system = sys.GetSystem()
	system.MainWindow.SetIcon(resourceGus2Png)

	file := filepath.Join(system.Storage, system.AppName) + ".log"
	var logger *os.File
	if logf, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666); err == nil {
		logger = logf
		defer func() {
			_ = logger.Close()
		}()
		log.SetOutput(logger)
	}
	log.Printf("fman - System:  %s\n", system)
	defer func() {
		system.Delete()
	}()
	//	preferences
	settings, err := sys.LoadUserPrefs(system.Storage, system.AppName)
	if err != nil {
		msg := fmt.Sprintf("Prefs Error. %s", err)
		n := fyne.NewNotification(sys.GetSystem().AppName, msg)
		sys.GetSystem().App.SendNotification(n)
		sys.Toast(msg, sys.ErrorToast)
		return
	}
	system.Settings = settings
	system.App.Settings().SetTheme(element.NewTheme(system.App.Preferences()))

	///// build the visual elements \\\\\

	aPanel, aBox := control.NewPanel("A")
	bPanel, bBox := control.NewPanel("B")
	aPanel.Twin = bPanel
	bPanel.Twin = aPanel

	// at top of each panel
	aTool := container.NewHBox(
		aPanel.Refresh, aPanel.MarkAll, aPanel.New, aPanel.Home,
		aPanel.Places, aPanel.Find, aPanel.History, aPanel.Copy,
		aPanel.Delete,
		aPanel.Source)
	bTool := container.NewHBox(
		bPanel.Source,
		bPanel.Refresh, bPanel.MarkAll, bPanel.New, bPanel.Home,
		bPanel.Places, bPanel.Find, bPanel.History, bPanel.Copy,
		bPanel.Delete)
	leftPane := container.NewBorder(aTool, nil, nil, nil, aBox)
	rightPane := container.NewBorder(bTool, nil, nil, nil, bBox)

	center := container.NewHSplit(leftPane, rightPane)
	bottom := newCommandBar(aPanel, bPanel)
	content := container.NewBorder(nil, bottom, nil, nil, center)

	// application cleanup
	system.MainWindow.SetOnClosed(func() {
		control.ClosePrefs()
		_ = sys.SavePrefs(system.Settings)
		app.CloseAppWindows()
	})
	system.MainWindow.SetContent(content)
	system.MainWindow.Resize(fyne.NewSize(800, 600))

	system.MainWindow.ShowAndRun()
	system.App.Quit()
}
