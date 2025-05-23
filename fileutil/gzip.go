package fileutil

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

/*

  File:    gzip.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*

 */

type GZipper struct {
	gzFile string
	reader *gzip.Reader
	target *os.File
}

func NewUnZGipper(gzFile string) (*GZipper, error) {
	f, err := os.Open(gzFile)
	if err != nil {
		return nil, err
	}
	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	z := GZipper{
		gzFile: gzFile,
		reader: r,
	}
	return &z, nil
}

func NewGZipper(gzFile string) (*GZipper, error) {
	target, err := os.Create(gzFile)
	if err != nil {
		return nil, err
	}
	z := GZipper{
		gzFile: gzFile,
		target: target,
	}
	return &z, nil
}
func (z *GZipper) Extract(dest string, notDone func()) (int, error) {
	defer func() {
		_ = z.reader.Close()
	}()
	x := filepath.Ext(z.gzFile)
	modTime := z.reader.Header.ModTime
	name := z.reader.Header.Name
	if name == "" {
		x = ".tgz"
	}
	switch strings.ToLower(x) {
	case ".tgz": // uncompress and extract in single operation
		fakeTar := &Tar{}
		fakeTar.reader = tar.NewReader(z.reader)
		return fakeTar.Extract(dest, notDone)
	default:
		switch strings.ToLower(filepath.Ext(name)) {
		case ".tar":
			fakeTar := &Tar{}
			fakeTar.reader = tar.NewReader(z.reader)
			return fakeTar.Extract(dest, notDone)
		default:
			_ = os.MkdirAll(dest, os.ModePerm)
			destination := filepath.Join(dest, name)
			log.Println("destination", destination, ", name", name, ", time", modTime)
			out, e := os.Create(destination)
			if e != nil {
				return 0, e
			}
			defer func() {
				_ = out.Close()
				_ = os.Chtimes(destination, modTime, modTime)
			}()
			b := make([]byte, 4096)
			n := 0
			for {
				notDone()
				nb, ee := z.reader.Read(b)
				if ee == io.EOF {
					break
				}
				n += nb
				nw, _ := out.Write(b[:nb])
				if nb != nw {
					break
				}
			}
			return 1, nil
		}
	}
}
func (z *GZipper) Compress(parent string, files []string, notDone func()) error {
	defer func() {
		_ = z.target.Close()
	}()
	archive := gzip.NewWriter(z.target)
	defer func() {
		_ = archive.Close()
	}()
	archive.Header.Comment = strings.ReplaceAll(z.gzFile, "\\", "/")
	switch runtime.GOOS {
	case "windows":
		archive.Header.OS = 0
	case "linux", "darwin", "freebsd", "netbsd", "openbsd", "solaris":
		archive.Header.OS = 3
	default:
		archive.Header.OS = 11
	}
	loc, _ := time.LoadLocation("Local")
	if len(files) == 1 {
		if len(parent) > 1 {
			parent += "/"
		}
		fileToGzip := files[0]
		archive.Header.Name = filepath.Base(fileToGzip)
		f, err := os.Open(fileToGzip)
		if err != nil {
			return err
		}
		info, err := f.Stat()
		if err != nil {
			return err
		}
		archive.Header.ModTime = info.ModTime().In(loc)
		b := make([]byte, 4096)
		for {
			nb, err := f.Read(b)
			if err != nil {
				return err
			}
			_, err = archive.Write(b[:nb])
			if err != nil {
				return err
			}
			if nb < 4096 {
				break
			}
		}
		return err
	}
	// use Tar
	archive.Header.Name = filepath.Base(parent) + ".tar"
	fakeTar := &Tar{}
	fakeTar.writer = tar.NewWriter(archive)
	return fakeTar.Compress(parent, files, notDone)
}
