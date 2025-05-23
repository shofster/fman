package app

/*

  File:    slideShow.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

  Display images.

*/

import (
	"fman/sys"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"os"
	"path/filepath"
)

var sliderCount = 1

func NewSlideShow(system *sys.System, path string) {

	// find all the images, based on extension
	// ix is the one double clicked on
	// start with it
	images, ix := buildImageList(system, path)

	image := canvas.NewImageFromFile(images[ix])
	image.FillMode = canvas.ImageFillContain
	c := container.New(layout.NewStackLayout(), image)

	label := widget.NewLabel(filepath.Base(path))
	label.Alignment = fyne.TextAlignCenter
	previous := widget.NewButton("", func() {
		ix--
		ix = showImage(c, images, ix, label)
	})
	previous.SetIcon(theme.NavigateBackIcon())
	next := widget.NewButton("", func() {
		ix++
		ix = showImage(c, images, ix, label)
	})
	next.SetIcon(theme.NavigateNextIcon())

	//hb := widget.NewHBox(previous, label, next)
	hb := container.New(layout.NewBorderLayout(nil, nil, previous, next),
		previous, next, label)

	content := container.New(layout.NewBorderLayout(nil, hb, nil, nil),
		hb, c)

	id := fmt.Sprintf("Slider(%d)", sliderCount)
	sliderCount++
	w := fyne.CurrentApp().NewWindow(id)
	openWindows[id] = w
	w.SetContent(content)
	w.Resize(fyne.NewSize(500, 380))
	w.SetFixedSize(false)
	w.Show()
}

func showImage(c *fyne.Container, images []string, nx int, label *widget.Label) int {
	if nx < 0 {
		nx = len(images) - 1
	}
	ix := nx % len(images)
	image := canvas.NewImageFromFile(images[ix])
	image.FillMode = canvas.ImageFillContain
	label.Text = filepath.Base(images[ix])
	c.Hide()
	c.Objects = nil
	c.Add(image)
	canvas.Refresh(c)
	label.Refresh()
	c.Show()
	return ix
}

// find all the images in the folder of the selected file
// return the index of it
func buildImageList(system *sys.System, path string) ([]string, int) {
	dir := filepath.Dir(path)
	name := filepath.Base(path)
	fi, _ := os.ReadDir(dir)
	var images []string
	ix := -1
	n := 0
	for _, file := range fi {
		if file.IsDir() {
			continue
		}
		switch sys.GetAssocType(system.Settings, file.Name()) {
		case "image":
			if file.Name() == name {
				ix = n
			}
			images = append(images, filepath.Join(dir, file.Name()))
			n++
		}
	}
	return images, ix
}
