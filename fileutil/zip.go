package fileutil

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/*

  File:    zip.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*

 */

type Zipper struct {
	zipFile string
	reader  *zip.ReadCloser
	target  *os.File
}

func NewUnZipper(zipFile string) (*Zipper, error) {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	z := Zipper{
		zipFile: zipFile,
		reader:  r,
	}
	return &z, nil
}

func NewZipper(zipFile string) (*Zipper, error) {
	target, err := os.Create(zipFile)
	if err != nil {
		return nil, err
	}
	z := Zipper{
		zipFile: zipFile,
		target:  target,
	}
	return &z, nil
}

func (z *Zipper) Extract(dest string, notDone func()) error {
	e := os.MkdirAll(dest, os.ModePerm)
	if e != nil {
		return e
	}
	var out *os.File
	defer func() {
		if out != nil {
			_ = out.Close()
		}
		z.reader = nil
	}()
	for _, file := range z.reader.File {
		notDone()
		isDir := file.FileInfo().IsDir()
		if isDir {
			dir := filepath.Dir(file.Name)
			path := filepath.Join(dest, dir)
			e := os.MkdirAll(path, os.ModePerm)
			if e != nil {
				return e
			}
			continue
		}
		// parent of this file
		parent := filepath.Dir(file.Name)
		if parent != "" && parent != "." {
			path := filepath.Join(dest, parent)
			em := os.MkdirAll(path, os.ModePerm)
			if em != nil {
				return em
			}
		}
		destination := filepath.Join(dest, file.Name)
		out, err := os.Create(destination)
		if err != nil {
			out = nil
			return err
		}
		fr, err := file.Open()
		if err != nil {
			return err
		}
		_, err = io.Copy(out, fr)
		if err != nil {
			return err
		}
		_ = os.Chtimes(destination, file.Modified, file.Modified)
		_ = out.Close()
		out = nil
	}
	return nil
}

func (z *Zipper) Compress(parent string, files []string, notDone func()) error {
	defer func() {
		_ = z.target.Close()
	}()
	if len(parent) > 1 {
		parent += "/"
	}
	archive := zip.NewWriter(z.target)
	defer func() {
		_ = archive.Close()
	}()
	for _, file := range files {
		notDone()
		err := addZipFile(archive, parent, file)
		if err != nil {
			return err
		}
	}
	return nil
}

// file is @ dir + file
func addZipFile(zipWriter *zip.Writer, parent string, file string) error {
	fileToZip, err := os.Open(file)
	if err != nil {
		return err
	}
	defer func() {
		_ = fileToZip.Close()
	}()
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}
	header, _ := zip.FileInfoHeader(info)
	path := file[len(parent):]
	if info.IsDir() {
		path += "/"
	}
	header.Name = strings.ReplaceAll(path, "\\", "/")
	loc, _ := time.LoadLocation("Local")
	header.Modified = header.Modified.In(loc)
	header.Method = zip.Deflate
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
