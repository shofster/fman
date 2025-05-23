//go:build windows

package fileutil

import (
	"context"
	"golang.org/x/sys/windows"
	"log"
	"os"
	"syscall"
	"time"
)

/*

  File:    windrives.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*
  Description: handle the discovery of Windows dirves A-Z
*/

func getDrives() (drives []string, err error) {

	var dll syscall.Handle
	dll, err = syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		log.Println("Error loading kernel32.dll", err)
		return
	}
	defer func(handle syscall.Handle) {
		_ = syscall.FreeLibrary(handle)
	}(dll)

	var logicalDrivesHandle uintptr
	logicalDrivesHandle, err = syscall.GetProcAddress(dll, "GetLogicalDrives")
	ret, _, err := syscall.SyscallN(logicalDrivesHandle, 0, 0, 0, 0)
	if err != syscall.Errno(0) {
		log.Println("Error calling GetLogicalDrives", err)
		return
	}
	err = nil

	done := make(chan bool)
	var stat error
	for i := 0; i < 26; i++ {

		if ret&1 == 1 {
			drive := string('A'+rune(i)) + ":\\"
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			go func() {
				_, stat = os.Stat(drive)
				done <- true
			}()
			select {
			case <-done:
				if stat == nil {
					drives = append(drives, drive)
				} else {
					err = stat
				}
			case <-ctx.Done():
				log.Println("Timeout on drive", drive)
			}
			cancel()
		}

		ret >>= 1
	}

	return
}

func getDiskUsage(vol string) (disk DiskUsage) {
	u16fname, err := syscall.UTF16FromString(vol)
	if err == nil {
		_ = windows.GetDiskFreeSpaceEx(&u16fname[0], &disk.Avail, &disk.All, &disk.Free)
		disk.Used = disk.All - disk.Free
	}
	return
}
