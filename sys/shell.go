package sys

import (
	"runtime"
	"strings"
)

/*

  File:    shell.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/

// var powerShell = [...]string{"cmd.exe", "/C", "start", "pwsh", "-NoExit", "-NoLogo", "-WorkingDirectory", "<DIR>"}
var powerShell = [...]string{"cmd.exe", "/C", "start", "pwsh", "-NoExit", "-NoLogo", "-WorkingDirectory", "<DIR>"}

// []string{"pwsh", "-NoExit", "-NoLogo", "-WorkingDirectory", "<DIR>", "-Command", "& { . $PROFILE }"}
var cmdShell = [...]string{"cmd.exe", "/K", "start", "/D", "<DIR>", "cmd.exe"}
var linuxShell = [...]string{"gnome-terminal", "--working-directory=<DIR>"}
var darwinShell = [...]string{"bash", "-c", "open -a Terminal \"<DIR>\""}
var defaultShell = [...]string{"sh", "term"}

func ShellCommand(dir string) []string {

	var g = func(params []string) []string {
		shell := make([]string, len(params))
		for i, v := range params {
			shell[i] = strings.Replace(v, "<DIR>", dir, -1)
		}
		return shell
	}

	switch runtime.GOOS {
	case "windows":

		if prefs.PowerShell {
			return g(powerShell[0:])
		}
		return g(cmdShell[0:])
	case "darwin":
		return g(darwinShell[0:])
	case "linux":
		return g(linuxShell[0:])
	default:
		return g(defaultShell[0:])
	}
}
