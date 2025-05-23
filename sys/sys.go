package sys

/*

  File:    sys.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

	Consolidate fman system information.

*/

import (
	"fman/element"
	"fman/misc"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"sync"
	"time"
)

/*

effective with Go 1.5

const goosList = "android darwin dragonfly freebsd linux nacl \
  netbsd openbsd plan9 solaris windows "

const goarchList = "386 amd64 amd64p32 arm arm64 ppc64 ppc64le \
   mips mipsle mips64 mips64le mips64p32 mips64p32le \ # (new)
   ppc s390 s390x sparc sparc64 "

*/

var system *System
var doOnce sync.Once

type System struct {
	misc.Sys
	AppName       string
	App           fyne.App
	Storage       string
	MainWindow    fyne.Window
	Settings      *Prefs
	Cline         *widget.Button
	Dir           string
	BusyIndicator *widget.ProgressBarInfinite
}

func GetSystem() *System {
	// Singleton System.
	doOnce.Do(func() {
		system = &System{AppName: "fman"}
		system.App = app.NewWithID("com.scsi.fman")
		system.Sys = misc.GetSys(system.AppName)
		system.Storage = system.App.Storage().RootURI().Path()
		system.MainWindow = system.App.NewWindow("File Manager (by Bob)")
	})
	return system
}

// "toString"
func (s System) String() string {
	return fmt.Sprintf("Application %s, ARCH %s, OS %s\n HOME %s\n STORAGE %s\n TEMP %s\n",
		s.AppName, s.ARtype, s.OStype, s.UserHome, s.Storage, s.TempDir)
}

type ToastType int

const (
	InfoToast = iota
	WarnToast
	ErrorToast
)

func Toast(txt string, toastType ToastType) {
	c := system.MainWindow.Canvas()
	s := c.Size()
	p := fyne.NewPos(s.Width, s.Height)
	var col color.NRGBA
	var duration time.Duration
	switch toastType {
	case InfoToast:
		duration = time.Millisecond + 500
		col = element.InfoColor()
	case WarnToast:
		duration = time.Second * 1
		col = color.NRGBA{R: 0xf0, G: 0x03, B: 0xfc, A: 0xff}
	case ErrorToast:
		duration = time.Second * 2
		col = element.FailColor()
	default:
		return
	}
	element.Toast(c, p, txt, col, duration)
}
