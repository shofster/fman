package app

import "fyne.io/fyne/v2"

/*

  File:    windows.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*
  Description: insure any new windows are closed on exit.
*/

var openWindows = make(map[string]fyne.Window)

func CloseAppWindows() {
	for _, w := range openWindows {
		w.Close()
	}
}
