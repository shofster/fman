package control

import (
	"fman/fileutil"
	"fman/sys"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

/*

  File:    filemgr.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*
   Functions to support local File operations in panel.go

*/

// CreatePath - create a file OR folder in the parent directory
func CreatePath(parent string, win *fyne.Window, cb func(string, string, error)) {
	selected := "File"
	choice := widget.NewRadioGroup([]string{selected, "Folder", "zip", "gzip", "tar"}, func(s string) {
		selected = s
	})
	choice.Resize(fyne.Size{Width: 150, Height: 40})

	label := widget.NewLabel("New Path:  ")
	entry := widget.NewEntry()
	entry.Wrapping = fyne.TextWrapOff
	entry.SetPlaceHolder(fmt.Sprintf("Relative to %s", parent))
	cc := widget.NewLabel("path name is required")
	cc.Hide()
	choice.OnChanged = func(s string) {
		selected = s
		cc.Show()
	}

	c := container.NewBorder(nil, cc, label, nil, label, entry)
	content := container.NewVBox(choice, c)

	title := fmt.Sprintf("New File/Folder in %s", parent)

	dialog.ShowCustomConfirm(title, "OK", "Cancel", content, func(v bool) {
		if v {
			name := entry.Text
			if name == "" {
				sys.Toast(" path name is required ! ", sys.ErrorToast)
			}
			path := filepath.Join(parent, entry.Text)
			switch selected {
			case "File":
				_, err := os.Stat(path)
				if err == nil {
					err = os.ErrExist
					cb(path, "", err)
					return
				}
				fi, err := os.Create(path)
				if err != nil {
					cb(path, "", os.ErrPermission)
					return
				}
				defer func() {
					_ = fi.Close()
				}()
				cb(path, "", nil)
			case "Folder":
				_, err := os.Stat(path)
				if err == nil {
					err = os.ErrExist
					cb(path, "", err)
					return
				}
				err = os.MkdirAll(path, os.ModePerm)
				if err == nil {
					_, err = os.Stat(path)
					if err == nil {
						cb(path, "", err)
						return
					}
				}
				cb(path, "", os.ErrPermission)
			default:
				cb(path, selected, nil)
			}
		}
	}, *win)
}

//
//////////////  file / folder copy(ies)  \\\\\\\\\\\\\\\\\\
//

// CopyModeSelect - query for copy parameters
func CopyModeSelect(source, destination string, count int, cb func(bool, bool, uint16), win *fyne.Window) {
	latest := widget.NewCheck("Latest ONLY", nil)
	all := widget.NewCheck("Overwrite ALL", nil)
	latest.SetChecked(true)
	latest.OnChanged = func(t bool) {
		if t {
			all.SetChecked(false)
		}
	}
	all.OnChanged = func(t bool) {
		if t {
			latest.SetChecked(false)
		}
	}
	choice := container.NewHBox(latest, all)
	label := widget.NewLabel("Buffer Size:  ")
	entry := widget.NewEntry()
	entry.SetPlaceHolder("8192")
	content := container.NewVBox(choice, container.NewHBox(label, entry))
	title := fmt.Sprintf("Copy %d files from %s to %s", count, source, destination)
	dialog.ShowCustomConfirm(title, "Continue", "Cancel", content, func(v bool) {
		var n uint16 = 8192
		if v {
			s := entry.Text
			if len(s) < 2 {
				s = "8192"
			}
			u, err := strconv.ParseUint(s, 10, 16)
			if err == nil {
				n = uint16(u)
			}
		}
		cb(v, all.Checked, n)

	}, *win)
}

func copyFile(source, destination string, all bool) error {
	destination = filepath.Join(destination, filepath.Base(source))
	infoS, _ := os.Stat(source)
	timeS := infoS.ModTime()
	infoD, errD := os.Stat(destination)
	// fail on any error but "does not exist" - that;s OK
	if errD != nil && !(os.IsExist(errD) || os.IsNotExist(errD)) {
		return errD
	}
	// checking dates on an existing, unless "all"
	if !all && (errD == nil) {
		timeD := infoD.ModTime()
		// skip if destination is not older than source
		if timeS.Before(timeD) || timeS.Equal(timeD) {
			return nil
		}
	}
	_, err := fileutil.CopyPlace(source, destination, timeS)
	return err
}

func IterateCopy(selected []string, destination string, all bool, bsize uint16, fl *FileLogger,
	refresh func(), done func(error)) {
	done(iterateCopy(selected, destination, all, bsize, fl, refresh))
}
func iterateCopy(selected []string, destination string, all bool, bsize uint16, fl *FileLogger, refresh func()) error {
	if len(selected) > 0 {
		for _, f := range selected {
			refresh()
			t, err := fileutil.GetPlaceType(f)
			if err != nil {
				return err
			}
			switch t {
			case fileutil.DirPlace:
				source := f
				currentDest := filepath.Join(destination, filepath.Base(f))
				err = os.MkdirAll(currentDest, os.ModePerm)
				if err != nil {
					return err
				}
				infoS, _ := os.Stat(source)
				timeS := infoS.ModTime()
				contents, _ := fileutil.PathContents(source)
				err = iterateCopy(contents, currentDest, all, bsize, fl, refresh)
				if err == nil {
					err = os.Chtimes(currentDest, timeS, timeS)
				}
			case fileutil.EmptyDirPlace:
				currentDest := filepath.Join(destination, filepath.Base(f))
				infoS, _ := os.Stat(f)
				timeS := infoS.ModTime()
				err = os.Chtimes(currentDest, timeS, timeS)
				err = os.MkdirAll(currentDest, os.ModePerm)
			case fileutil.FilePlace:
				for {
					err = copyFile(f, destination, all)
					if err == nil {
						fl.Console.Speak(filepath.Join(destination, filepath.Base(f)))
						fl.FileCount++
						break
					}
					action := fl.Error(err)
					if action == ActionAbort {
						break
					}
					if action == ActionSkip {
						err = nil
						break
					}
				}
			case fileutil.OtherPlace:
				log.Printf("Unable to copy %s\n", filepath.Join(destination, filepath.Base(f)))
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
