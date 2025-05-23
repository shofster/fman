package fileutil

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

/*

  File:    places.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*
  Description:
*/

// LoadPlaces enumerates the logical drives / standard places on the system.
func LoadPlaces() (drives []string) {
	switch runtime.GOOS {
	case "windows", "linux":
		drives, _ = getDrives()
	default:
		drives = append(drives, "/")
	}
	return
}

type DiskUsage struct {
	All   uint64
	Used  uint64
	Free  uint64
	Avail uint64
}

func PrettyDiskSize(s uint64) string {
	var B uint64 = 1073741824
	if s > B {
		return fmt.Sprintf("%4.1fGB", float32(s)/float32(B))
	}
	B /= 1024
	if s > B {
		return fmt.Sprintf("%4.1fMB", float32(s/1024./B))
	}
	B /= 1024
	if s > B {
		return fmt.Sprintf("%4.1fKB", float32(s/1024./1024./B))
	}
	return fmt.Sprintf("%db", s)
}

func GetDiskUsage(filename string) DiskUsage {
	switch runtime.GOOS {
	case "windows", "linux":
		return getDiskUsage(filename)
	default:
		return DiskUsage{}
	}
}

// TooManyFilesError is a custom error when trying to display 3000 files bogs down
// error is given, but first 256 file names are returned
type TooManyFilesError struct {
	Path  string
	Count int
}

func (e *TooManyFilesError) Error() string {
	return fmt.Sprintf("%d", e.Count)
}

// PathContents gets the names of all files in a directory/
func PathContents(path string) ([]string, error) {
	_, x := os.Stat(path)
	if os.IsNotExist(x) {
		return nil, errors.New("sys.PathContents: " + path + "," + x.Error())
	}
	fi, err := os.ReadDir(path)
	p := make([]string, 0, len(fi))
	for _, file := range fi {
		p = append(p, filepath.Join(path, file.Name()))
	}
	return p, err
}

// DirTreeList gets (recursive) names of all files and directories.
func DirTreeList(path string, f func(string) error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error in DirTreeList", r)
		}
	}()
	_ = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		return f(p)
	})
}

type PlaceType int

const (
	EmptyDirPlace = iota
	FilePlace
	DirPlace
	OtherPlace
)

// a "toString" of the PlaceType
func (t PlaceType) String() string {
	return [...]string{"EmptyDir", "File", "Dir", "Other"}[t]
}

// GetPlaceType return a limited PlaceType.
func GetPlaceType(place string) (PlaceType, error) {
	fi, err := os.Stat(place)
	if err != nil {
		return OtherPlace, err
	}
	if fi.IsDir() {
		sub, ep := PathContents(place)
		if ep != nil {
			return DirPlace, ep
		}
		if len(sub) == 0 {
			return EmptyDirPlace, nil
		} else {
			return DirPlace, nil
		}
	}
	if fi.Mode().IsRegular() {
		return FilePlace, nil
	}
	return OtherPlace, nil
}

// CopyPlace copies a source file from the destination directory.
// and reset the new file's time to the original.
func CopyPlace(source, destination string, time time.Time) (uint64, error) {
	in, err := os.Open(source)
	if err != nil {
		return 0, err
	}
	defer func(in *os.File) {
		if e := in.Close(); e != nil {
		}
	}(in)
	out, err1 := os.Create(destination)
	if err1 != nil {
		return 0, err1
	}
	defer func(out *os.File) {
		if e := out.Close(); e != nil {
			log.Println("close copy close", e, time)
		}
		_ = os.Chtimes(destination, time, time)
	}(out)
	nBytes, err2 := io.Copy(out, in)
	if err2 != nil {
		return uint64(nBytes), err2
	}
	return uint64(nBytes), err2
}
