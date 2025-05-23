package app

import (
	"fman/element"
	"fman/sys"
	"fmt"
	"fyne.io/fyne/v2"
	"log"
	"os"
	"path/filepath"
)

/*

  File:    viewer.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*
  Description:
*/

var viewerCount = 1

func NewViewer(system *sys.System, path string) {
	reader, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return
	}
	defer func(reader *os.File) {
		_ = reader.Close()
	}(reader)
	tv := element.NewTextViewer(system.MainWindow, filepath.Base(path), reader, nil, 2)
	id := fmt.Sprintf("Viewer(%d)", viewerCount)
	viewerCount++
	w := fyne.CurrentApp().NewWindow(id)
	w.SetContent(tv.Content)
	w.Resize(fyne.NewSize(600, 400))
	openWindows[id] = w
	w.SetFixedSize(false)
	w.Show()
}
