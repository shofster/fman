package app

/*

  File:    finder.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

  Find files in the current directory tree that match criteria.

*/

import (
	"errors"
	"fman/fileutil"
	"fman/sys"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"io"
	"regexp"
	"strconv"
)

type Finder struct {
	system   *sys.System
	id       string
	parent   string
	selected func(string)
	win      fyne.Window
	cancel   bool
	scroll   *container.Scroll
	box      *fyne.Container
	criteria string
	max      int
}

func NewFinder(system *sys.System, id string, selected func(string)) *Finder {
	finder := &Finder{
		system:   system,
		id:       id,
		selected: selected}
	buildWindow(finder)
	return finder
}
func buildWindow(finder *Finder) {
	finder.cancel = false
	finder.win = fyne.CurrentApp().NewWindow("Finder in Panel " + finder.id)
	//finder.win.SetOnClosed(func() {
	//	finder.win = nil
	//})
	finder.win.SetCloseIntercept(func() { finder.win.Hide() })
	criteria := widget.NewEntry()
	criteria.PlaceHolder = "Enter Search Criteria"
	sliderValue := widget.NewLabel(" 10")
	finder.max = 10
	slider := widget.NewSlider(10, 100)
	slider.Value = 10
	slider.Step = 10
	slider.OnChanged = func(v float64) {
		finder.max = int(v)
		sliderValue.SetText(fmt.Sprintf("%3s", strconv.Itoa(finder.max)))
	}
	stop := widget.NewButtonWithIcon("", theme.MediaStopIcon(), nil)
	stop.OnTapped = func() {
		stop.Disable()
		finder.cancel = true
	}
	stop.Disable()
	dismiss := widget.NewButton(" Dismiss ", func() {
		stop.Disable()
		finder.cancel = true
		finder.win.Hide()
	})
	criteria.Validator = func(s string) error {
		return nil
	}
	criteria.ActionItem = widget.NewButtonWithIcon(
		"", theme.SearchIcon(),
		func() {
			if criteria.Text == "" {
				return
			}
			if _, err := regexp.Compile("(?i)" + criteria.Text); err != nil {
				criteria.SetValidationError(errors.New("invalid"))
				return
			}
			finder.criteria = criteria.Text
			stop.Enable()
			find(finder)
		})
	criteria.OnSubmitted = func(k string) {
		if criteria.Text == "" {
			return
		}
		if _, err := regexp.Compile("(?i)" + k); err != nil {
			t := fmt.Sprintf("%s : %s", criteria.Text, err.Error())
			criteria.SetText(t)
			return
		}
		finder.criteria = k
		stop.Enable()
		find(finder)
	}
	controls := container.NewHBox(sliderValue, slider, stop, dismiss)
	params := container.NewBorder(nil, nil, nil, controls, criteria)
	finder.box = container.NewVBox()
	finder.scroll = container.NewVScroll(finder.box)
	results := container.NewStack(finder.scroll)
	content := container.NewBorder(params, nil, nil, nil, results)
	finder.win.SetContent(content)
	finder.win.Resize(fyne.NewSize(600, 300))
	finder.win.SetFixedSize(false)
	finder.win.Canvas().Focus(criteria)
}

func find(finder *Finder) {
	finder.cancel = false
	finder.box.Objects = make([]fyne.CanvasObject, 0)
	finder.box.Refresh()
	rx := regexp.MustCompile("(?i)" + finder.criteria)
	finder.system.BusyIndicator.Start()
	go search(finder, rx)
}

func search(finder *Finder, regexp *regexp.Regexp) {
	n := 0
	// executes until cancel, max, or error
	fileutil.DirTreeList(finder.parent, func(path string) error {
		if finder.cancel {
			return io.EOF
		}
		finder.system.BusyIndicator.Refresh()
		if regexp.MatchString(path) {
			n++
			finder.box.Objects = append(finder.box.Objects,
				newFindItem(finder, path))
			finder.box.Refresh()
			if n >= finder.max {
				return io.EOF
			}
		}
		return nil
	})
	finder.box.Objects = append(finder.box.Objects,
		newFindItem(finder, fmt.Sprintf("Finished Searching ... found %d items", n)))
	finder.system.BusyIndicator.Stop()
	go finder.box.Refresh()
}

type FindItem struct {
	widget.Label
	finder   *Finder
	onTapped func(string)
}

var _ fyne.Focusable = (*FindItem)(nil)

func (t *FindItem) FocusGained() {
}
func (t *FindItem) FocusLost() {
}
func (t *FindItem) Focused() bool {
	return false
}
func (t *FindItem) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyHome:
		t.finder.scroll.ScrollToTop()
	case fyne.KeyEnd:
		t.finder.scroll.ScrollToBottom()
	}
}
func (t *FindItem) TypedRune(rune) {
}

func newFindItem(finder *Finder, display string) *FindItem {
	fi := &FindItem{}
	fi.ExtendBaseWidget(fi)
	fi.SetText(display)
	fi.finder = finder
	fi.onTapped = finder.selected
	return fi
}

func (t *FindItem) Tapped(_ *fyne.PointEvent) {
	t.onTapped(t.Text)
}

// Find - new search criteria
func Find(finder *Finder, parent string) {
	finder.win.SetTitle(fmt.Sprintf("Finding in %s", parent))
	finder.win.Show()
	finder.cancel = false
	finder.parent = parent
	finder.win.Show()
}
