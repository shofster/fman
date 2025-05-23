package control

/*

  File:    ops.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

    Major operations (New, Trash ...) on a Panel.

*/

import (
	"errors"
	"fman/app"
	"fman/fileutil"
	"fman/sys"
	"fmt"
	"fyne.io/fyne/v2"
	"os"
	"path/filepath"
	"strings"
)

func PanelRefresh(panel *Panel) {
	if panel.parent != "" {
		buildNewItems(panel, panel.parent)
	}
	panel.UpdateFavorites()
}
func PanelNew(panel *Panel) {
	if panel.parent == "" {
		sys.Toast("Select a Place", sys.WarnToast)
		PanelRefresh(panel)
		return
	}
	// get extension and apply it
	CreatePath(panel.parent, &sys.GetSystem().MainWindow,
		func(path, ext string, err error) { // extension selected callback
			//			log.Printf("path is %s, zipext %s, extenstion is %s", path, ext, fext)
			// CreatePath will create normal files and folders
			if ext == "" { // not zip or gzip
				if err != nil && !errors.Is(err, os.ErrExist) {
					msg := fmt.Sprintf("Create Error. %s", err)
					sys.Toast(msg, sys.ErrorToast)
				} else {
					sys.Toast("Created "+path, sys.InfoToast)
				}
				PanelRefresh(panel)
				return
			}

			selected := panel.Twin.dir.GetSelected()
			if len(selected) < 1 || panel.Twin.parent == "" {
				sys.Toast("No Source(s) Selected", sys.WarnToast)
				return
			}

			sys.GetSystem().BusyIndicator.Start()
			defer func() {
				sys.GetSystem().BusyIndicator.Stop()
			}()

			files := make([]string, 0)
			for _, file := range selected {
				p := filepath.Join(panel.Twin.parent, file.DisplayName())
				fi, _ := os.Stat(p)
				if !fi.IsDir() {
					files = append(files, p)
				}
			}
			for _, file := range selected {
				p := filepath.Join(panel.Twin.parent, file.DisplayName())
				fi, _ := os.Stat(p)
				if fi.IsDir() {
					// collect ALL full file names
					fileutil.DirTreeList(p, func(s string) error {
						files = append(files, s)
						return nil
					})
				}
			}
			_ = len(files)
			fext := filepath.Ext(path)
			// log.Printf("fext %s, ext is %s, count is %d, path is %s ", fext, ext, nfiles, path)

			switch ext {
			case "zip":
				if fext == "" {
					path += ".zip"
				}
				z, err := fileutil.NewZipper(path)
				if err == nil {
					err = z.Compress(panel.Twin.parent, files, func() {
						sys.GetSystem().BusyIndicator.Refresh()
					})
				}
				if err != nil {
					sys.Toast(fmt.Sprintf("Create Zip Fail %s", err), sys.ErrorToast)
					break
				}
				PanelRefresh(panel)
			case "tar":
				if fext == "" {
					path += ".tar"
				}
				z, err := fileutil.NewTar(path)
				if err == nil {
					err = z.Compress(panel.Twin.parent, files, func() {
						sys.GetSystem().BusyIndicator.Refresh()
					})
				}
				if err != nil {
					sys.Toast(fmt.Sprintf("Create Tar Fail %s", err), sys.ErrorToast)
					break
				}
				PanelRefresh(panel)
			case "gzip":
				if fext == "" { // didn't supply an extension
					path += ".gz"
				}
				z, err := fileutil.NewGZipper(path)
				if err == nil {
					err = z.Compress(panel.Twin.parent, files, func() {
						sys.GetSystem().BusyIndicator.Refresh()
					})
				}
				if err != nil {
					sys.Toast(fmt.Sprintf("Create Gzip Fail %s", err), sys.ErrorToast)
					break
				}
				PanelRefresh(panel)
			}
		})
}
func panelHome(panel *Panel) {
	p := sys.GetSystem().UserHome
	panelPlace(panel, p)
}
func panelFind(panel *Panel) {
	app.Find(panel.Finder, panel.parent)
}

// Place and Copy are called from the Panel's select widgets and buttons
func panelPlace(panel *Panel, h string) {
	if h == "* TEMP *" {
		h = sys.GetSystem().TempDir
	}
	panel.History.Options = sys.Remove(panel.History.Options, h)
	panel.History.Options = append([]string{h}, panel.History.Options...)
	if len(panel.History.Options) > 10 {
		panel.History.Options = panel.History.Options[0:10]
	}
	buildItems(panel, h)
	sys.GetSystem().Settings.AddHistory(h)
}

func panelCopy(panel *Panel) {
	if panel.Twin.parent == "" {
		sys.Toast("No Destination Selected", sys.WarnToast)
		return
	}
	if panel.parent == panel.Twin.parent {
		sys.Toast("Copying Into Same Folder", sys.InfoToast)
	}
	selected := panel.dir.GetSelected()
	if len(selected) < 1 || panel.Twin.parent == "" {
		sys.Toast("File(s) and Destination Must Be Selected", sys.WarnToast)
		return
	}
	var names []string
	for _, file := range selected {
		names = append([]string{filepath.Join(panel.parent, file.DisplayName())}, names...)
	}
	CopyModeSelect(panel.parent, panel.Twin.parent, len(names),
		func(cont, all bool, size uint16) {
			if !cont {
				return
			}
			sys.GetSystem().BusyIndicator.Start()
			fl := NewFileLogger()
			fl.Console.Speak(fmt.Sprintf("Copy from %s\n to %s\n", panel.parent, panel.Twin.parent))
			go IterateCopy(names[0:], panel.Twin.parent, all, size, fl, func() {
				sys.GetSystem().BusyIndicator.Refresh()
			}, func(err error) {
				sys.GetSystem().BusyIndicator.Stop()
				if err != nil {
					fl.Error(err)
					msg := fmt.Sprintf("Copy Error. %s", err)
					n := fyne.NewNotification(sys.GetSystem().AppName, msg)
					sys.GetSystem().App.SendNotification(n)
					sys.Toast(msg, sys.ErrorToast)
				}
				fl.Done(fl.FileCount)
				fl.Close()
				PanelRefresh(panel.Twin)
			})
		}, &sys.GetSystem().MainWindow)
}
func executeView(path string) {
	app.NewViewer(sys.GetSystem(), path)
}
func executeEdit(path string) {
	cmd := sys.GetSystem().Settings.GetEditCommand(path)
	if cmd != nil {
		err := cmd.Start()
		if err != nil {
			msg := fmt.Sprintf("Edit Error. %s", err)
			n := fyne.NewNotification(sys.GetSystem().AppName, msg)
			sys.GetSystem().App.SendNotification(n)
			sys.Toast(msg, sys.ErrorToast)
		}
	}
}
func panelDelete(panel *Panel) {
	selected := panel.dir.GetSelected()
	if len(selected) < 1 {
		sys.Toast("No Files(s) Selected", sys.WarnToast)
		return
	}
	msg := fmt.Sprintf("%d Files / Folders", len(selected))
	fileutil.VerifyDelete(sys.GetSystem().MainWindow, msg, func(yes bool) {
		if !yes {
			return
		}
		for _, s := range selected {
			path := filepath.Join(panel.parent, s.DisplayName())
			err := os.RemoveAll(path)
			if err != nil {
				sys.Toast(fmt.Sprintf("Fail %s on file %s, Delete Terminated", err.Error(), path), sys.ErrorToast)
				break
			}
		}
		PanelRefresh(panel)
	})
}
func panelAction(panel *Panel, path string) {
	sys.GetSystem().BusyIndicator.Start()
	defer func() {
		sys.GetSystem().BusyIndicator.Stop()
	}()
	// see if known (by file extension) special handling
	switch sys.GetAssocType(sys.GetSystem().Settings, path) {
	case "image":
		app.NewSlideShow(sys.GetSystem(), path)
		return
	case "zip": // unzip
		dest := filepath.Join(sys.GetSystem().TempDir,
			strings.Replace(filepath.Base(path), ".", "_", -1))
		z, _ := fileutil.NewUnZipper(path)
		ez := z.Extract(dest, func() {
			sys.GetSystem().BusyIndicator.Refresh()
		})
		if ez != nil {
			sys.Toast(fmt.Sprintf("Fail %s on file %s, UNZIP Terminated", ez.Error(), path), sys.ErrorToast)
		}
		panelPlace(panel, dest)
		return
	case "gzip": // ungzip
		dest := filepath.Join(sys.GetSystem().TempDir,
			strings.Replace(filepath.Base(path), ".", "_", -1))
		z, _ := fileutil.NewUnZGipper(path)
		_, ez := z.Extract(dest, func() {
			sys.GetSystem().BusyIndicator.Refresh()
		})
		if ez != nil {
			sys.Toast(fmt.Sprintf("Fail %s on file %s, UNGZIP Terminated", ez.Error(), path), sys.ErrorToast)
		}
		panelPlace(panel, dest)
		return
	case "tar":
		dest := filepath.Join(sys.GetSystem().TempDir,
			strings.Replace(filepath.Base(path), ".", "_", -1))
		z, _ := fileutil.NewUnTar(path)
		_, ez := z.Extract(dest, func() {
			sys.GetSystem().BusyIndicator.Refresh()
		})
		if ez != nil {
			sys.Toast(fmt.Sprintf("Fail %s on file %s, UNTAR Terminated", ez.Error(), path), sys.ErrorToast)
		}
		panelPlace(panel, dest)
		return
	}
	// not known, try as command
	////executeCommand(path, make([]string, 0))
	_ = fileutil.Browse(path)
}
