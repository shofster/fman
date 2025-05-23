package control

/*

  File:    panel.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

    One of two file manager panels.

*/

import (
	"fman/app"
	"fman/fileutil"
	"fman/sys"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var activePanel *Panel

type Panel struct {
	id     string
	canvas fyne.Canvas
	// the Box containing the visual items, and the actual Items
	box *fyne.Container
	// top is the current path and goto parent button
	top     *fyne.Container
	listBox *fyne.Container
	current *widget.Label
	upItem  *widget.Button
	up      *widget.Label
	list    *fileutil.CustomList
	dir     *fileutil.DirectoryEntry
	// full (local) path to the parent of the displayed files
	// "" means nothing showing
	parent string
	// last selected index
	selected int
	// secondary selected
	secondarySelect fileutil.FileEntry

	// controls set or managed by others
	Refresh *widget.Button
	New     *widget.Button
	Home    *widget.Button
	Places  *widget.Select
	History *widget.Select
	Find    *widget.Button
	Copy    *widget.Button
	Delete  *widget.Button
	Source  *widget.Button
	MarkAll *widget.Button
	Finder  *app.Finder
	Popup   *widget.PopUpMenu
	Twin    *Panel
}

// "toString"
func (p *Panel) String() string {
	return fmt.Sprintf("id %s, parent %s", p.id, p.parent)
}

func NewPanel(id string) (*Panel, *fyne.Container) {
	panel := &Panel{
		id:       id,
		parent:   "",
		selected: -1,
		canvas:   sys.GetSystem().MainWindow.Canvas(),
	}

	// first 2 items in the panel - current base name, and parent
	panel.current = widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	panel.upItem = widget.NewButton("", func() {
		activePanel = panel
		panelPlace(panel, filepath.Dir(panel.parent))
	})
	panel.upItem.SetIcon(theme.MoveUpIcon())
	panel.top = container.NewVBox(panel.current, panel.upItem)
	// item 3 is the list of file base names
	panel.showEmpty() // create List
	panel.listBox = container.NewStack(panel.list)
	panel.box = container.NewBorder(panel.top, nil, nil, nil, panel.listBox)

	// 2 Selects for favorites and history
	panel.Places = widget.NewSelect(fileutil.LoadPlaces(), func(value string) {
		activePanel = panel
	})
	panel.Places.PlaceHolder = " Places "
	panel.Places.OnChanged = func(value string) {
		panel.Places.Selected = panel.Places.PlaceHolder
		activePanel = panel
		panelPlace(panel, value)
	}

	// operation Buttons
	panel.Refresh = widget.NewButton("", func() {
		activePanel = panel
		PanelRefresh(panel)
		if panel.selected != -1 {
			panel.UpdateFavorites()
			panel.list.ScrollTo(panel.selected + 2)
			panel.list.Refresh()

		}
	})
	panel.Refresh.SetIcon(theme.ContentRedoIcon())
	panel.New = widget.NewButton("", func() {
		activePanel = panel
		PanelNew(panel)
	})
	panel.New.SetIcon(theme.ContentAddIcon())
	panel.Home = widget.NewButton("", func() {
		activePanel = panel
		panelHome(panel)
	})
	panel.Home.SetIcon(theme.HomeIcon())
	// invoke the finder app in this directory
	panel.Find = widget.NewButton("", func() {
		activePanel = panel
		panelFind(panel)
	})
	panel.Find.SetIcon(theme.SearchIcon())
	panel.Finder = app.NewFinder(sys.GetSystem(), panel.id, func(path string) {
		activePanel = panel
		pt, _ := fileutil.GetPlaceType(path)
		switch pt {
		case fileutil.FilePlace:
			panelPlace(panel, filepath.Dir(path))
		case fileutil.DirPlace:
			panelPlace(panel, path)
		default:
		}
	})
	panel.Copy = widget.NewButton("", func() {
		activePanel = panel
		panelCopy(panel)
	})
	panel.Copy.SetIcon(theme.ContentCopyIcon())

	view := fyne.NewMenuItem("Text Viewer", func() {
		if panel.secondarySelect.IsDir() {
			sys.Toast(fmt.Sprintf("%s is a Directory", panel.secondarySelect.DisplayName()), sys.WarnToast)
			return
		}
		executeView(filepath.Join(panel.secondarySelect.Name()))
	})
	edit := fyne.NewMenuItem("Text Editor", func() {
		if panel.secondarySelect.IsDir() {
			sys.Toast(fmt.Sprintf("%s is a Directory", panel.secondarySelect.DisplayName()), sys.WarnToast)
			return
		}
		executeEdit(filepath.Join(panel.secondarySelect.Name()))
	})
	props := fyne.NewMenuItem("Properties Editor", func() {
		//if panel.secondarySelect.IsDir() {
		//	sys.Toast(fmt.Sprintf("%s is a Directory", panel.secondarySelect.DisplayName()), sys.WarnToast)
		//	return
		//}
		app.FileInfoEdit(sys.GetSystem().MainWindow, panel.secondarySelect.Name())
	})
	menu := fyne.NewMenu("File Options", view, edit, props)
	panel.Popup = widget.NewPopUpMenu(menu, panel.canvas)

	panel.Delete = widget.NewButton("", func() {
		activePanel = panel
		panelDelete(panel)
	})
	panel.Delete.SetIcon(theme.DeleteIcon())

	panel.Source = widget.NewButton("", func() {
	})
	if panel.id == "A" {
		panel.Source.SetIcon(theme.MediaFastForwardIcon())
	} else {
		panel.Source.SetIcon(theme.MediaFastRewindIcon())
	}
	panel.Source.OnTapped = func() {
		activePanel = panel
		panelPlace(panel.Twin, panel.dir.Name())
	}

	panel.MarkAll = widget.NewButton("", func() {
	})
	panel.MarkAll.SetIcon(theme.ConfirmIcon())
	panel.MarkAll.OnTapped = func() {
		activePanel = panel
		m := panel.list.Route[fileutil.CtrlA]
		m(fileutil.CtrlA)
	}

	panel.History = widget.NewSelect(sys.GetSystem().Settings.GetHistory(), func(value string) {
		activePanel = panel
	})
	panel.History.PlaceHolder = " Visited "
	panel.History.OnChanged = func(value string) {
		panel.History.Selected = panel.History.PlaceHolder
		panelPlace(panel, value)
	}

	panel.disableOperations()
	activePanel = panel
	panel.UpdateFavorites()
	panel.box.Refresh()
	return panel, panel.box
}
func (p *Panel) disableOperations() {
	p.Refresh.Disable()
	p.New.Disable()
	p.upItem.Disable()
	p.Find.Disable()
	p.Copy.Disable()
	p.Delete.Disable()
	p.Source.Disable()
	p.MarkAll.Disable()
}
func (p *Panel) enableOperations() {
	p.Refresh.Enable()
	p.New.Enable()
	p.upItem.Enable()
	p.Find.Enable()
	p.Copy.Enable()
	p.Delete.Enable()
	p.Source.Enable()
	p.MarkAll.Enable()
}

// UpdateFavorites is called when first loaded to set the saved Favorites
func (p *Panel) UpdateFavorites() {
	p.Places.Options = fileutil.LoadPlaces()
	p.Places.Options = append(p.Places.Options, "* TEMP *")
	p.Places.Options = append(sys.GetSystem().Settings.Favorites, p.Places.Options...)
}

// first line of table - not tappable
func (p *Panel) showCurrent() {
	p.enableOperations()
	p.current.SetText(fmt.Sprintf("%s", filepath.Base(p.parent)))
}

// "previous" - second line of table - tappable
func (p *Panel) showPrevious() {
	p.upItem.SetText(filepath.Dir(p.parent))
}
func (p *Panel) showEmpty() {
	p.current.SetText("Select Home, Place, OR Visited")
	p.upItem.Text = ""
	p.parent = ""
	p.list = fileutil.NewCustomList()
	p.list.Length = func() int {
		return 1
	}
	p.list.CreateItem = func() fyne.CanvasObject {
		return widget.NewIcon(theme.FileIcon())
	}
	p.list.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {}
}

// error condition
func (p *Panel) showError(err error) {
	p.disableOperations()
	e := err.Error()
	if len(e) > 50 {
		e = e[len(e)-50:]
	}
	p.current.SetText(fmt.Sprintf("%s", e))
	p.parent = "ERROR"
	p.upItem.SetText(p.parent)
	p.list.Length = func() int {
		return 1
	}
	p.list.CreateItem = func() fyne.CanvasObject {
		return widget.NewIcon(theme.FileIcon())
	}
	p.list.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {}
	p.listBox.Objects[0] = p.list
}

// ///////////////////////////

func buildItems(panel *Panel, newPlace string) {
	sys.GetSystem().Dir = newPlace
	if newPlace != panel.parent && newPlace != "" {
		panel.selected = -1
		buildNewItems(panel, newPlace)
	}
}
func buildNewItems(panel *Panel, newPlace string) {
	fs := fileutil.FileSelectFilter{
		FileType:   fileutil.Any,
		Multiple:   true,
		Date:       sys.GetSystem().Settings.DateTime,
		Descending: sys.GetSystem().Settings.Descending,
		DtFormat:   sys.GetSystem().Settings.DateTimeFormat,
		Canvas:     panel.canvas}
	if !sys.GetSystem().Settings.Hidden {
		fs.Hidden = sys.GetSystem().Settings.HiddenFiles
	}
	fa := fileutil.FileSelectAction{
		OnClick: func(e fileutil.FileEntry) {
			activePanel = panel
			panel.selected = e.Index()
		},
		OnSecondaryClick: func(entry fileutil.FileEntry, event *fyne.PointEvent) {
			activePanel = panel
			panel.secondarySelect = entry
			panel.Popup.ShowAtPosition(event.AbsolutePosition)
		},
		OnDoubleClick: func(file fileutil.FileEntry) {
			activePanel = panel
			i, _ := file.Info()
			m := i.Mode().String()
			if strings.HasPrefix(m, "L") {
				l, _ := os.Readlink(file.Name())
				log.Printf("%s is link to %s\n", file.Name(), l)
				return
			}
			if file.IsDir() {
				sys.GetSystem().Dir = newPlace
				panelPlace(panel, filepath.Join(panel.parent, file.DisplayName()))
			} else {
				panelAction(panel, filepath.Join(panel.parent, file.DisplayName()))
			}
		},
	}
	var err error
	panel.list, panel.dir, err = fileutil.NewFileList(newPlace, fs, fa,
		nil)
	if err != nil {
		panel.showError(err)
		return
	}
	panel.parent = newPlace
	panel.showCurrent()
	panel.showPrevious()
	panel.listBox.Objects[0] = panel.list
	panel.listBox.Refresh()
}
