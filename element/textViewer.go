package element

/*

  File:    textViewer.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.
*/
/*
  Description: simple text viewer with search ability.
*/

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"io"
	"log"
	"regexp"
	"strings"
)

type TextViewer struct {
	Content *fyne.Container
	Err     error
	name    string
	view    *widget.TextGrid
	nRows   int
}
type searcher struct {
	find          string
	isRegEx       bool
	caseSensitive bool
}
type match struct {
	row  int
	col1 int
	col2 int
}

func NewTextViewer(window fyne.Window, name string, reader io.Reader, buttons []*widget.Button,
	tabSize int) *TextViewer {

	viewer := TextViewer{
		Content: nil,
		name:    name,
		view:    widget.NewTextGrid(),
	}
	scroll := container.NewScroll(viewer.view)

	title := widget.NewLabel(name)
	title.Importance = widget.HighImportance
	title.Alignment = fyne.TextAlignTrailing

	buttonBar := container.NewHBox(layout.NewSpacer())
	lineNumbers := widget.NewCheck("line #s", nil)
	whiteSpace := widget.NewCheck("white space", nil)
	buttonBar = container.NewHBox()
	buttonBar.Objects = append(buttonBar.Objects, lineNumbers)
	buttonBar.Objects = append(buttonBar.Objects, whiteSpace)
	var matches []match
	find := widget.NewButtonWithIcon("SEARCH...", theme.SearchIcon(), func() {
		n := 0
		for _, match := range matches {
			viewer.highLight(match.row, match.col1, match.col2, false)
			n++
		}
		if n > 0 {
			viewer.view.Refresh()
		}
		viewer.getSearchParameters(window, func(s searcher) {
			found := viewer.markFinds(s)
			matches = found
			if len(matches) < 1 {
				dialog.ShowInformation("No Matches", fmt.Sprintf("for '%s'", s.find), window)
			}
			viewer.view.Refresh()
		})
	})
	buttonBar.Objects = append(buttonBar.Objects, find)

	// add in any user buttons
	if len(buttons) > 0 {
		buttonBar.Objects = append(buttonBar.Objects, layout.NewSpacer())
		for _, button := range buttons {
			buttonBar.Objects = append(buttonBar.Objects, button)
		}
	}

	viewer.view.ShowLineNumbers = true
	lineNumbers.SetChecked(viewer.view.ShowLineNumbers)
	lineNumbers.OnChanged = func(show bool) {
		viewer.view.ShowLineNumbers = show
		viewer.view.Refresh()
	}
	viewer.view.ShowWhitespace = false // use spaces in place of \t, \r,...
	whiteSpace.SetChecked(viewer.view.ShowWhitespace)
	whiteSpace.OnChanged = func(show bool) {
		viewer.view.ShowWhitespace = show
		viewer.view.Refresh()
	}
	viewer.view.TabWidth = tabSize

	viewer.Content = container.NewBorder(title, buttonBar, nil, nil, scroll)

	content, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("Error %s reading: %s\n", err, name)
		viewer.Err = err
		content = []byte(fmt.Sprintf("io.ReadAll failure on %s\n", name))
	}
	viewer.view.SetText(string(content))
	viewer.nRows = len(viewer.view.Rows) - 1 // skip empty last line

	// last line
	line := viewer.getRowAsString(viewer.nRows - 1)
	log.Println("name", name, "rows:", viewer.nRows, ", last:", line,
		"len", len(line))

	return &viewer
}

// highLight sets normal or reverse colors for a range of columns on a roe
func (v *TextViewer) highLight(row, col1, col2 int, rev bool) {
	tgs := &widget.CustomTextGridStyle{
		TextStyle: widget.TextGridStyleDefault.Style(),
		BGColor:   theme.Color(theme.ColorNameBackground),
		FGColor:   theme.Color(theme.ColorNameForeground),
	}
	if rev {
		tgs.BGColor = theme.Color(theme.ColorNameForeground)
		tgs.FGColor = theme.Color(theme.ColorNameBackground)
	}
	v.view.SetStyleRange(row, col1, row, col2, tgs)
}

// getRowAsString fetches the UTF8 characters from a row
func (v *TextViewer) getRowAsString(rowId int) (utf8 string) {
	gridRowCells := v.view.Row(rowId).Cells // runes
	for _, r := range gridRowCells {
		utf8 += string(r.Rune)
	}
	return
}

// getSearchParameters - a form dialog to get user search criteria
func (v *TextViewer) getSearchParameters(window fyne.Window, cb func(s searcher)) {
	var params searcher

	var items = make([]*widget.FormItem, 0)

	entry := widget.NewEntry()
	entry.PlaceHolder = "3 minimum"
	find := widget.NewFormItem("Text to find. (minimum 3 chars)", entry)
	isRegExp := widget.NewCheck("", func(b bool) {
		params.isRegEx = b
	})
	regExp := widget.NewFormItem("Regular Expession?", isRegExp)
	isCaseSensitive := widget.NewCheck("", func(b bool) {
		params.caseSensitive = b
	})
	caseSensitive := widget.NewFormItem("Case Sensitive?", isCaseSensitive)
	spacer := widget.NewFormItem("", widget.NewLabel("                           "))

	items = append(items, spacer)
	items = append(items, find)
	items = append(items, regExp)
	items = append(items, caseSensitive)

	entry.Validator = func(str string) error {
		if len(str) < 3 {
			return errors.New("find string must be > 2 characters")
		}
		if params.isRegEx { // Check doesn't support validation
			re, err := regexp.Compile(str)
			if err != nil || re == nil {
				return errors.New("invalid regular expression")
			}
		}
		return nil
	}

	dialog.ShowForm("Search Criteria for "+v.name, "search", "cancel", items, func(b bool) {
		if b {
			params.find = entry.Text
			cb(params)
		}
	}, window)

	return
}

func (v *TextViewer) markFinds(s searcher) (matches []match) {

	find := s.find
	var re = regexp.MustCompile(find)
	if !s.caseSensitive {
		find = strings.ToLower(find)
	}

	for r := 0; r < v.nRows; r++ {
		row := v.getRowAsString(r)
		m := match{
			row: -1,
		}
		if s.isRegEx { // (?) prefix for case insensitive
			cols := re.FindStringIndex(row)
			if cols != nil {
				m.row = r
				m.col1 = cols[0]
				m.col2 = cols[1]
			}
		} else {
			if !s.caseSensitive {
				row = strings.ToLower(row)
			}
			m.col1 = strings.Index(row, find)
			if m.col1 > -1 {
				m.row = r
				m.col2 = m.col1 + len(find) - 1
			}
		}
		if m.row > -1 {
			v.highLight(m.row, m.col1, m.col2, true)
			matches = append(matches, m)
		}
	}
	return
}
