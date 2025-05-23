package sys

/*

  File:    assoc.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

    Manage associations between file extensions and processing type

*/

import (
	"strings"
)

func setDefaultAssociations(p *Prefs) {
	a := createZipAssoc()
	p.Assoc = append(p.Assoc, a)
	b := createGzipAssoc()
	p.Assoc = append(p.Assoc, b)
	a = createTarAssoc()
	p.Assoc = append(p.Assoc, a)
	c := createBatchAssoc()
	p.Assoc = append(p.Assoc, c)
	d := createImageAssoc()
	p.Assoc = append(p.Assoc, d)
}

// add any new associations to an old Prefs
func checkNewAssociations(p *Prefs) {
	checkZipAssoc(p)
	checkGzipAssoc(p)
	checkTarAssoc(p)
	checkBatchAssoc(p)
	checkImageAssoc(p)
}

func GetAssocType(p *Prefs, name string) string {
	ext := strings.ToLower(name)
	for _, a := range p.Assoc {
		for _, e := range a.Extensions {
			if strings.HasSuffix(ext, e) {
				return a.Type
			}
		}
	}
	return ""
}

func checkZipAssoc(p *Prefs) {
	for _, t := range p.Assoc {
		if t.Type == "zip" {
			return
		}
	}
	a := createZipAssoc()
	p.Assoc = append(p.Assoc, a)
}
func createZipAssoc() FileAssoc {
	a := FileAssoc{}
	a.Type = "zip"
	a.Extensions = make([]string, 0)
	a.Extensions = append(a.Extensions, ".zip")
	a.Extensions = append(a.Extensions, ".jar")
	a.Extensions = append(a.Extensions, ".war")
	a.Extensions = append(a.Extensions, ".ear")
	return a
}

func checkGzipAssoc(p *Prefs) {
	for _, t := range p.Assoc {
		if t.Type == "gzip" {
			return
		}
	}
	b := createGzipAssoc()
	p.Assoc = append(p.Assoc, b)
}
func createGzipAssoc() FileAssoc {
	b := FileAssoc{}
	b.Type = "gzip"
	b.Extensions = make([]string, 0)
	b.Extensions = append(b.Extensions, ".tgz")
	b.Extensions = append(b.Extensions, ".gz")
	b.Extensions = append(b.Extensions, ".gzip")
	return b
}

func checkTarAssoc(p *Prefs) {
	for _, t := range p.Assoc {
		if t.Type == "tar" {
			return
		}
	}
	b := createGzipAssoc()
	p.Assoc = append(p.Assoc, b)
}
func createTarAssoc() FileAssoc {
	b := FileAssoc{}
	b.Type = "tar"
	b.Extensions = make([]string, 0)
	b.Extensions = append(b.Extensions, ".tar")
	return b
}

func checkBatchAssoc(p *Prefs) {
	for _, t := range p.Assoc {
		if t.Type == "batch" {
			return
		}
	}
	c := createBatchAssoc()
	p.Assoc = append(p.Assoc, c)
}
func createBatchAssoc() FileAssoc {
	c := FileAssoc{}
	c.Type = "batch"
	c.Extensions = make([]string, 0)
	c.Extensions = append(c.Extensions, ".bat")
	c.Extensions = append(c.Extensions, ".cmd")
	return c
}

func checkImageAssoc(p *Prefs) {
	for _, t := range p.Assoc {
		if t.Type == "image" {
			return
		}
	}
	d := createImageAssoc()
	p.Assoc = append(p.Assoc, d)
}
func createImageAssoc() FileAssoc {
	d := FileAssoc{}
	d.Type = "image"
	d.Extensions = make([]string, 0)
	d.Extensions = append(d.Extensions, ".png")
	d.Extensions = append(d.Extensions, ".jpg")
	d.Extensions = append(d.Extensions, ".jpeg")
	return d
}
