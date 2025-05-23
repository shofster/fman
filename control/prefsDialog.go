package control

/*

  File:    prefsDialog.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

    Pop up dialog to manage the Settings in fman.json


*/

import (
	"fman/fileutil"
	"fman/sys"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"log"
	_ "strings"
)

var prefsWindow fyne.Window
var favorites *widget.Select
var removeFavorite string
var remove *widget.Button
var browserSel = fileutil.FileSelectFilter{
	Title:      "Select Browser Location",
	FileType:   fileutil.File,
	FileSelect: fileutil.Open,
	Multiple:   false,
	Hidden:     ""}
var favSel = fileutil.FileSelectFilter{
	Title:      "Select Favorite Folder",
	FileType:   fileutil.Dir,
	FileSelect: fileutil.Open,
	Multiple:   false,
	Hidden:     ""}
var lastPopup *widget.PopUp

func newPrefsDialog(system *sys.System, panelA *Panel, panelB *Panel) {
	prefsWindow = fyne.CurrentApp().NewWindow("Preferences")
	dateFormat := widget.NewSelect(fileutil.RFCtypes, func(_ string) {})
	dateFormat.PlaceHolder = system.Settings.DateTimeFormat
	hidden := widget.NewCheck("", func(bool) {
	})
	hidden.SetChecked(system.Settings.Hidden)
	hiddenFiles := widget.NewEntry()
	hiddenFiles.Text = system.Settings.HiddenFiles
	browser := widget.NewButton("", nil)
	browser.OnTapped = func() {
		lastPopup = fileutil.FileSelect(browserSel, nil, prefsWindow, func(files []string) {
			lastPopup = nil
			if len(files) > 0 {
				browser.SetText(system.Settings.Browser)
			}
		})
	}
	browser.SetText(system.Settings.Browser)
	form := &widget.Form{
		Items: []*widget.FormItem{},
		OnSubmit: func() { // handle form submission
			sys.GetSystem().Settings.SetDateTimeFormat(dateFormat.Selected)
			sys.GetSystem().Settings.SetHiddenFiles(hiddenFiles.Text)
			sys.GetSystem().Settings.SetHidden(hidden.Checked)
			sys.GetSystem().Settings.SetBrowser(browser.Text)
			err := sys.SavePrefs(system.Settings)
			if err != nil {
				log.Printf("Save Settings FAILED: %v\n", err)
			}
			panelA.UpdateFavorites()
			panelB.UpdateFavorites()
			prefsWindow.Hide()
			_ = sys.SavePrefs(sys.GetSystem().Settings)
			showingPrefsDialog = false
		},
		OnCancel: func() { // handle form cancel
			prefsWindow.Hide()
			showingPrefsDialog = false
		},
	}
	form.SubmitText = "Save & Close"
	remove = widget.NewButton("<Remove Old>", func() {
		l := sys.Remove(favorites.Options, removeFavorite)
		favorites.Options = l
		favorites.Hide()
		favorites.Selected = "<Add New>"
		favorites.Show()
		remove.Hide()
		sys.GetSystem().Settings.RemoveFavorite(removeFavorite)
	})
	remove.Hide()
	favorites = widget.NewSelect([]string{"<Add New>"}, func(value string) {
		if value == "<Add New>" {
			lastPopup = fileutil.FileSelect(favSel, nil, prefsWindow, func(dirs []string) {
				favorites.Hide()
				lastPopup = nil
				if len(dirs) > 0 {
					d := dirs[0]
					favorites.Options = sys.Remove(favorites.Options, d)
					favorites.Options = append(favorites.Options, d)
					sys.GetSystem().Settings.SetFavorites(favorites.Options[1:])
					favorites.Selected = dirs[0]
				}
				favorites.Show()
			})
		} else {
			removeFavorite = value
			remove.Hide()
			remove.Text = value
			remove.Show()
		}
	})
	favorites.Options = append(favorites.Options, system.Settings.Favorites...)
	spacer := widget.NewButton("", func() {
	})
	spacer.Hide()
	form.Append("", spacer)
	// preferred browser
	form.Append("Browser", browser)
	// format for Date and Time
	form.Append("Date & Time Format", dateFormat)
	// ignore "hidden" files
	form.Append("Show Hidden Files", hidden)
	// the reg exp to identify files to be "hidden"
	form.Append("Hidden Files", hiddenFiles)
	// list of Mounts
	form.Append("Favorites", favorites)
	form.Append("Remove", remove)

	form.Append("", spacer)
	form.Append("", spacer)
	form.Append("", spacer)
	form.Append("", spacer)
	prefsWindow.SetContent(form)
	prefsWindow.SetOnClosed(func() {
		if prefsWindow != nil {
			if lastPopup != nil {
				lastPopup.Hide()
				lastPopup = nil
			}
			showingPrefsDialog = false
			prefsWindow = nil
		}
	})
}

var showingPrefsDialog bool

func ShowAndEditPrefs(system *sys.System, panelA *Panel, panelB *Panel) {
	if !showingPrefsDialog {
		if prefsWindow == nil {
			newPrefsDialog(system, panelA, panelB)
		}
		prefsWindow.Show()
		showingPrefsDialog = true
	}
}

// ClosePrefs called when app is closing
func ClosePrefs() {
	if prefsWindow != nil {
		if lastPopup != nil {
			lastPopup.Hide()
			lastPopup = nil
		}
		prefsWindow.Close()
		prefsWindow = nil
		showingPrefsDialog = false
	}
}
