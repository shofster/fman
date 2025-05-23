package app

/*

  File:    fileinfo.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*
  Description:
*/

import (
	"errors"
	"fman/sys"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var DefaultDateTimeFormat = "02 Jan 06 15:04 MST"

func FileInfoEdit(window fyne.Window, path string) {
	fi, err := os.Lstat(path)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	sys.GetSystem().BusyIndicator.Start()
	defer func() {
		sys.GetSystem().BusyIndicator.Stop()
	}()
	mode := fmt.Sprintf("%s", fi.Mode())
	items := make([]*widget.FormItem, 0)
	name := fi.Name()
	newName := name
	dir := "File"
	p := message.NewPrinter(language.English)
	size := p.Sprintf("%d bytes", fi.Size())
	if fi.IsDir() {
		dir = "Directory"
		var bytes int64
		var files int
		var visit = func(p string, f os.FileInfo, err error) error {
			i, err := os.Stat(p)
			if err == nil {
				files++
				bytes += i.Size()
				if files > 50000 {
					s := "too many files - abort"
					sys.Toast(s, sys.ErrorToast)
					return errors.New(s)
				}
			}
			return nil
		}
		_ = filepath.Walk(path, visit)
		size = p.Sprintf("%d files, %d bytes", files, bytes)
	} else if strings.HasPrefix(mode, "L") {
		dir = "Link"
	}
	dt := fi.ModTime().Format(DefaultDateTimeFormat)
	modified := dt
	items = append(items, &widget.FormItem{
		Text: fmt.Sprintf("Info:"),
		Widget: widget.NewLabel(fmt.Sprintf("%s\n%s\n%s\n%s\n%s      ",
			name, dir, mode, size, dt)),
	})
	nameV := widget.NewEntryWithData(binding.BindString(&newName))
	nameV.Wrapping = fyne.TextWrapOff
	nameV.SetText(name)
	items = append(items, &widget.FormItem{
		Text:   "New Name:",
		Widget: nameV,
	})
	nameV.Validator = func(s string) error {
		if len(s) < 1 {
			return errors.New("name is required ")
		}
		return nil
	}
	dtV := widget.NewEntryWithData(binding.BindString(&modified))
	nameV.Wrapping = fyne.TextWrapOff
	dtV.SetText(dt)
	dtV.Validator = func(s string) error {
		_, err := time.Parse(DefaultDateTimeFormat, s)
		return err
	}
	items = append(items, &widget.FormItem{
		Text:   "Modified:",
		Widget: dtV,
	})

	dlg := dialog.NewForm(path, "Apply", "Cancel", items,
		func(b bool) {
			if b {
				if newName != name {
					d := filepath.Dir(path)
					err := os.Rename(path, filepath.Join(d, newName))
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
				}
				if dt != modified {
					t, err := time.Parse(DefaultDateTimeFormat, modified)
					if err == nil {
						err = os.Chtimes(path, t, t)
						if err != nil {
							dialog.ShowError(err, window)
						}
					} else {
						dialog.ShowError(err, window)
					}
				}
			}
		},
		window)
	dlg.Show()

}
