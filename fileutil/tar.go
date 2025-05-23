package fileutil

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/*

  File:    tar.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*

 */

type Tar struct {
	tarFile string
	reader  *tar.Reader
	target  *os.File
	writer  *tar.Writer
}

func NewUnTar(tarFile string) (*Tar, error) {
	f, err := os.Open(tarFile)
	if err != nil {
		return nil, err
	}
	r := tar.NewReader(f)
	z := Tar{
		tarFile: tarFile,
		reader:  r,
	}
	return &z, nil
}

func NewTar(tarFile string) (*Tar, error) {
	target, err := os.Create(tarFile)
	if err != nil {
		return nil, err
	}
	archive := tar.NewWriter(target)
	z := Tar{
		tarFile: tarFile,
		target:  target,
		writer:  archive,
	}
	return &z, nil
}

func (z *Tar) Extract(dest string, notDone func()) (int, error) {
	_ = os.MkdirAll(dest, os.ModePerm)
	defer func() {
		z.reader = nil
	}()

	var out *os.File
	defer func() {
		if out != nil {
			_ = out.Close()
		}
	}()
	count := 0
	for {
		notDone()
		header, err := z.reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}
		destination := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(destination, os.ModePerm)
			if err != nil {
				return count, err
			}
		case tar.TypeReg:
			err = os.MkdirAll(filepath.Dir(destination), os.ModePerm)
			out, err = os.Create(destination)
			if err != nil {
				out = nil
				return count, err
			}
			_, err = io.Copy(out, z.reader)
			if err != nil {
				return count, err
			}
			_ = os.Chtimes(destination, header.ModTime, header.ModTime)
			_ = out.Close()
			count++
			out = nil
		}
	}
	return count, nil
}

func (z *Tar) Compress(parent string, files []string, notDone func()) error {
	if len(parent) > 1 {
		parent += "/"
	}
	defer func() {
		_ = z.writer.Close()
	}()
	defer func() {
		_ = z.target.Close()
	}()
	for _, file := range files {
		notDone()
		err := addTarFile(z.writer, parent, file)
		if err != nil {
			return err
		}
	}
	return nil
}
func addTarFile(tarWriter *tar.Writer, parent string, file string) error {
	// input file
	fileToTar, err := os.Open(file)
	if err != nil {
		return err
	}
	defer func() {
		_ = fileToTar.Close()
	}()
	info, err := fileToTar.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	header := &tar.Header{
		Size:     info.Size(),
		Mode:     int64(info.Mode()),
		ModTime:  info.ModTime(),
		Typeflag: tar.TypeReg,
	}
	if info.IsDir() {
		header.Typeflag = tar.TypeDir
	}
	path := file[len(parent):]
	header.Name = strings.ReplaceAll(path, "\\", "/")
	loc, _ := time.LoadLocation("Local")
	header.ModTime = header.ModTime.In(loc)
	err = tarWriter.WriteHeader(header)
	if err == nil {
		_, err = io.Copy(tarWriter, fileToTar)
	}
	return nil
}
