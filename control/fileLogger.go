package control

import (
	"fman/element"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

/*

  File:    fileLogger.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*
  Description: window to track file operations
*/

type FileLogger struct {
	win       fyne.Window
	Console   *element.Console
	DirCount  int
	FileCount int
	Action    chan int
	IgnoreAll bool
}

const (
	ActionRetry = iota
	ActionSkip
	ActionSkipAll
	ActionAbort
	ActionDone = -1
)

func NewFileLogger() *FileLogger {
	fl := FileLogger{}
	fl.Action = make(chan int)
	buttons := make([]*widget.Button, 5)
	buttons[0] = widget.NewButton("Retry", func() { fl.Action <- ActionRetry })
	buttons[1] = widget.NewButton("Skip", func() { fl.Action <- ActionSkip })
	buttons[2] = widget.NewButton("Ignore All", func() { fl.Action <- ActionSkipAll })
	buttons[3] = widget.NewButton("Abort", func() { fl.Action <- ActionAbort })
	buttons[4] = widget.NewButton("Done", func() { fl.Action <- ActionDone })
	buttons[0].Disable()
	buttons[1].Disable()
	buttons[2].Disable()
	buttons[3].Disable()
	buttons[4].Hide()
	fl.win = fyne.CurrentApp().NewWindow("FileLogger")
	fl.Console = element.NewConsole(fl.win, element.Prompt, buttons, 20)
	fl.win.SetContent(container.NewStack(fl.Console.Content))
	fl.win.Resize(fyne.NewSize(300, 300))
	fl.win.SetFixedSize(false)
	fl.win.Show()
	return &fl
}
func (fl *FileLogger) Error(err error) (action int) {
	if fl.IgnoreAll {
		action = ActionSkip
		return
	}
	fl.Console.Buttons[0].Enable()
	fl.Console.Buttons[1].Enable()
	fl.Console.Buttons[2].Enable()
	fl.Console.Buttons[3].Enable()
	fl.Console.Speak(fmt.Sprintf("Error:\n%s\nSelect Action", err))
	action = <-fl.Action
	if action == ActionSkipAll {
		fl.IgnoreAll = true
		action = ActionSkip
	}
	fl.Console.Buttons[0].Disable()
	fl.Console.Buttons[1].Disable()
	fl.Console.Buttons[2].Disable()
	fl.Console.Buttons[3].Disable()
	return
}
func (fl *FileLogger) Done(count int) {
	fl.Console.Buttons[0].Hide()
	fl.Console.Buttons[1].Hide()
	fl.Console.Buttons[2].Hide()
	fl.Console.Buttons[3].Hide()
	fl.Console.Buttons[4].Show()
	fl.Console.Content.Refresh()
	fl.Console.Speak(fmt.Sprintf("Done:\n%d files\n", count))
	_ = <-fl.Action
	return
}
func (fl *FileLogger) Close() {
	fl.win.Close()
}
