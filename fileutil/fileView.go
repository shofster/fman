package fileutil

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"net/url"
)

/*

  File:    fileView.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*

 */

type ProtocolType int

const (
	FILE ProtocolType = iota
	ZIP
	TAR
	GZIP
)

var extMap = map[string]ProtocolType{
	".ZIP":  ZIP,
	".JAR":  ZIP,
	".WAR":  ZIP,
	".EAR":  ZIP,
	".TAR":  TAR,
	".GZ":   GZIP,
	".GZIP": GZIP,
	".TGZ":  GZIP,
}

type FileSelectType int

const (
	Any FileSelectType = iota
	File
	Dir
)

type FileSelectOp int

const (
	Open FileSelectOp = iota
	Save
)

type ListIconType int

type FileSelectFilter struct {
	Canvas     fyne.Canvas
	Title      string
	FileType   FileSelectType
	FileSelect FileSelectOp
	Multiple   bool
	Hidden     string
	Ext        string
	Date       bool
	DtFormat   string
	Descending bool
}

type FileSelectAction struct {
	OnClick          func(file FileEntry)
	OnDoubleClick    func(FileEntry)
	OnSecondaryClick func(entry FileEntry, event *fyne.PointEvent)
}
type CustomList struct {
	widget.List
	Route map[fyne.Shortcut]func(fyne.Shortcut)
}

func NewCustomList() *CustomList {
	cl := &CustomList{}
	cl.ExtendBaseWidget(cl)
	cl.Route = make(map[fyne.Shortcut]func(fyne.Shortcut))
	return cl
}
func (cl *CustomList) TypedShortcut(s fyne.Shortcut) {
	if _, ok := s.(*desktop.CustomShortcut); !ok {
		cl.TypedShortcut(s)
		return
	}
	if v, ok := cl.Route[s]; ok {
		v(s)
	}
}

var CtrlA = &desktop.CustomShortcut{KeyName: fyne.KeyA, Modifier: fyne.KeyModifierControl}

type fileView interface {
	Open(path string, sel FileSelectFilter) (*DirectoryEntry, error)
	Close()
}

func VerifyDelete(appWindow fyne.Window, content string, res func(bool)) {
	dialog.ShowConfirm("Is it OK to completely Remove?", content, res, appWindow)
}

func VerifyOverwrite(appWindow fyne.Window, content string, res func(bool)) {
	dialog.ShowConfirm("Is it OK to Replace?", content, res, appWindow)
}
func AskFolder(appWindow fyne.Window, loc string, res func(string)) {
	var file = ""
	result := binding.BindString(&file)
	items := make([]*widget.FormItem, 0)
	but := widget.NewButton(loc+"                                         .", func() {})
	but.Disable()
	items = append(items, widget.NewFormItem("Location", but))
	items = append(items, widget.NewFormItem("Name", widget.NewEntryWithData(result)))
	dialog.ShowForm("Create a NEW Folder", "OK", "CANCEL", items, func(ok bool) {
		if !ok {
			res("")
		} else {
			res(file)
		}
	}, appWindow)
}

func urlFromString(str string) (*url.URL, error) {
	u, err := url.Parse(str)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Browse sends a file to the App
func Browse(file string) error {
	u, err := urlFromString(file)
	if err != nil {
		return err
	}
	return fyne.CurrentApp().OpenURL(u)
}
