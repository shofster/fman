package misc

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

/*

  File:    sys.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*
  Description:
*/

type Sys struct {
	ARtype   string
	OStype   string
	UserHome string
	AppDir   string
	TempDir  string
}

var system Sys
var doOnce sync.Once

// GetSys gets system wide (global) parameters.
func GetSys(name string) Sys {
	// Singleton Sys.
	doOnce.Do(func() {
		system = Sys{OStype: runtime.GOOS, ARtype: runtime.GOARCH}
		system.UserHome = userHomeDir()
		system.AppDir = appDataDir(name)
		system.TempDir = tempDataDir(name)
	})
	return system
}

// "toString"
func (sys Sys) String() string {
	return fmt.Sprintf("ARCH %s, OS %s, HOME %s, APPDATA %s, Temp %s", sys.ARtype,
		sys.OStype, sys.UserHome, sys.AppDir, sys.TempDir)
}

// Delete the temp files amd directory
func (sys Sys) Delete() {
	_ = os.RemoveAll(system.TempDir)
	_ = os.Remove(system.TempDir)
}

// Drives enumerates the logical drives / standard places on the system.
func (sys Sys) Drives() (drives []string, err error) {
	switch sys.OStype {
	case "windows":
		for _, d := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			drive := string(d) + ":\\"
			f, err := os.Open(drive)
			if err == nil {
				drives = append(drives, drive)
				_ = f.Close()
			}
		}
	case "linux":
		dirs := []string{"/", "/etc", "/home", "/media", "/mnt", "/tmp", "/usr"}
		for _, dir := range dirs {
			if f, err := os.Open(dir); err == nil {
				drives = append(drives, dir)
				_ = f.Close()
			}
		}
	default:
		drives = append(drives, "/")
	}
	return
}

// userHomeDir provides the path to the user's Home directory (windows or other).
func userHomeDir() string {
	home, err := os.UserHomeDir()
	if err == nil {
		return home
	}
	home = os.Getenv("HOMEDRIVE")
	if home == "" {
		panic("Unable to get User Home")
	}
	return home
}

// appDataDir provides the path to the user's Application directory for (windows or other).
func appDataDir(name string) string {
	if configDir, e1 := os.UserConfigDir(); e1 == nil {
		p := filepath.Join(configDir, name)
		_, e2 := os.Stat(p)
		if e2 == nil {
			return p
		} else if os.IsNotExist(e2) {
			if os.MkdirAll(p, os.ModeDir) == nil {
				return p
			}
		}
	}
	p := filepath.Join(userHomeDir(), name)
	if os.MkdirAll(p, os.ModeDir) == nil {
		return p
	}
	panic("Unable to get User AppDataDir")
}

// tempDataDir provides the path to a Temp directory for (windows or other).
func tempDataDir(name string) string {
	tempDir, err := os.MkdirTemp("", name+"*")
	if err == nil {
		return tempDir
	}
	return filepath.Join(userHomeDir(), name, "_temp")
}

/*
func listWindowsPlaces() []string {
	var drives []string
	for _, d := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		drive := string(d) + ":\\"
		f, err := os.Open(drive)
		if err == nil {
			drives = append(drives, drive)
			_ = f.Close()
		}
	}
	return drives
}
*/
