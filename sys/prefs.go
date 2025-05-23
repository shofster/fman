package sys

/*

  File:    prefs.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

    Manage the persistent User Preferences / Settings

*/

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2/widget"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var prefs *Prefs

// these are the widgets
// for both Panels to use (as needed)
// configured and saved in prefs.go

type FileAssoc struct {
	Type       string   `json:"type"`
	Extensions []string `json:"ext"`
}
type Prefs struct {
	DateTimeFormat string      `json:"dtformat"`
	Hidden         bool        `json:"hidden"`
	HiddenFiles    string      `json:"hiddenfiles"`
	DateTime       bool        `json:"date"`
	Descending     bool        `json:"descending"`
	Ask            bool        `json:"ask"`
	History        []string    `json:"history"`
	Favorites      []string    `json:"favorites"`
	Edit           []string    `json:"edit"`
	Assoc          []FileAssoc `json:"assoc"`
	Batch          []string    `json:"batch"`
	Browser        string      `json:"browser"`
	Browse         []string    `json:"browse"`
	Text           int         `json:"text"`
	Font           int         `json:"font"`
	Path           string
	hidden         *widget.Check
	monospace      *widget.Check
	dateTime       *widget.Check
	descending     *widget.Check
}

func (p *Prefs) SetHidden(t bool) {
	p.hidden.SetChecked(t)
}
func (p *Prefs) SetMonospace(t bool) {
	p.monospace.SetChecked(t)
}
func (p *Prefs) GetDateTimeFormat() string {
	return p.DateTimeFormat
}
func (p *Prefs) SetDateTimeFormat(format string) {
	p.DateTimeFormat = format
}
func (p *Prefs) GetHiddenFiles() string {
	return p.HiddenFiles
}
func (p *Prefs) SetHiddenFiles(regexp string) {
	p.HiddenFiles = regexp
}
func (p *Prefs) GetBrowser() string {
	return p.Browser
}
func (p *Prefs) SetBrowser(browser string) {
	p.Browser = browser
}
func (p *Prefs) SetFavorites(favorites []string) {
	p.Favorites = favorites
}
func (p *Prefs) RemoveFavorite(favorite string) {
	p.Favorites = Remove(p.Favorites, favorite)
}
func (p *Prefs) GetBatchCommand(path string) *exec.Cmd {
	batch := make([]string, len(p.Batch))
	copy(batch, p.Batch[0:len(p.Batch)])
	ix := Index(batch, "<PATH>")
	if ix != -1 {
		batch[ix] = path
	}
	cmd := exec.Command(batch[0], batch[1:]...)
	return cmd
}
func (p *Prefs) GetShellCommand() *exec.Cmd {
	if system.Dir == "" {
		return nil
	}
	return getCurrentShellCommand(system.Dir)
}
func getCurrentShellCommand(dir string) *exec.Cmd {
	shell := ShellCommand(dir)
	log.Println("shell", shell)
	cmd := exec.Command(shell[0], shell[1:]...)
	cmd.Dir = dir
	return cmd
}
func (p *Prefs) GetEditCommand(file string) *exec.Cmd {
	shell := make([]string, len(p.Edit))
	for i, v := range p.Edit {
		shell[i] = strings.Replace(v, "<FILE>", file, -1)
	}
	cmd := exec.Command(shell[0], shell[1:]...)
	return cmd
}
func (p *Prefs) GetBrowseCommand(file string) *exec.Cmd {
	browse := make([]string, len(p.Edit))
	copy(browse, p.Browse[0:len(p.Browse)])
	ix := Index(browse, "<BROWSER>")
	if ix != -1 {
		browse[ix] = filepath.Base(p.Browser)
	}
	ix = Index(browse, "<URL>")
	if ix != -1 {
		browse[ix] = file
	}
	cmd := exec.Command(p.Browser, browse[1:]...)
	//	cmd.Dir = filepath.Dir(p.Browser)
	return cmd
}

// AddHistory is dynamic and updates every change of Directory
// in either Panel
func (p *Prefs) AddHistory(h string) {
	if h == "." || strings.Contains(h, system.TempDir) {
		return
	}
	p.History = Remove(p.History, h)
	p.History = append([]string{h}, p.History...)
	if len(p.History) > 10 {
		p.History = p.History[0:10]
	}
}
func (p *Prefs) GetHistory() []string {
	var history []string
	for _, h := range p.History {
		history = append(history, h)
	}
	return history
}

func LoadUserPrefs(storage, appName string) (*Prefs, error) {
	return LoadPrefs(filepath.Join(storage, appName+".json"))
}

// LoadPrefs load the last instance of Preferences.
// create one if first execution by a user
func LoadPrefs(path string) (*Prefs, error) {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return NewPrefs(path), nil
	}
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("sys.LoadPrefs: " + err.Error())
	}

	p := NewPrefs(path)
	err = json.Unmarshal(jsonData, &p)
	if err != nil {
		return nil, errors.New("sys.LoadPrefs: " + err.Error())
	}
	p.hidden.SetChecked(p.Hidden)
	p.dateTime.SetChecked(p.DateTime)
	p.descending.SetChecked(p.Descending)
	p.Path = path
	if p.History == nil {
		p.History = make([]string, 0)
	}
	if p.Favorites == nil {
		p.Favorites = make([]string, 0)
	}

	if p.Assoc == nil || len(p.Assoc) < 1 {
		p.Assoc = make([]FileAssoc, 0)
		setDefaultAssociations(p)
	}
	// insure any newly coded one's get Defaults
	checkNewAssociations(p)

	// log.Printf("Edit: %v\n", p.Edit)
	if p.Edit == nil || len(p.Edit) < 1 {
		setDefaultEditCmd(p)
	}
	// log.Printf("Batch: %v\n", p.Batch)
	if p.Batch == nil || len(p.Batch) < 1 {
		setDefaultBatchCmd(p)
	}
	// log.Printf("Browse: %v %v\n", p.Browser, p.Browse)
	if p.Browser == "" || p.Browse == nil || len(p.Browse) < 1 {
		setDefaultBrowsing(p)
	}
	return p, err
}

func SavePrefs(prefs *Prefs) error {
	if prefs == nil {
		return nil
	}
	b, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return errors.New("sys.SavePrefs: " + err.Error())
	}
	// create or truncate
	f, err := os.Create(prefs.Path)
	if err != nil {
		return errors.New("sys.SavePrefs: " + err.Error())
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = f.Write(b)
	if err != nil {
		return errors.New("sys.SavePrefs: " + err.Error())
	}
	return nil
}

////////////////////////////////////////////////

// these are the widgets
// for both Panels to use (as needed) for file(s) display
func widgets(p *Prefs) {
	// hidden (true) will display hidden Files
	p.hidden = widget.NewCheck("Hidden", func(t bool) {
		p.Hidden = t
	})
	// dateTime (true) will sort by DataTime, instead of Name
	p.dateTime = widget.NewCheck("Date/Time", func(t bool) {
		p.DateTime = t
	})
	// descending (true) will sort the DataTime or Name in descenging order
	p.descending = widget.NewCheck("Descending", func(t bool) {
		p.Descending = t
	})

}
func GetHidden(p *Prefs) *widget.Check {
	return p.hidden
}
func GetDateTime(p *Prefs) *widget.Check {
	return p.dateTime
}
func GetDescending(p *Prefs) *widget.Check {
	return p.descending
}

// remove the time zone code, if don't want / need
var defaultDateTimeFormat = "01/02/06 15:04"
var defaultHiddenFiles = "(^[^\\w].+)|(.+\\.bak)$"
var defaultWinCmdShell = [...]string{"cmd.exe", "/K", "start", "/D", "<DIR>", "cmd.exe"}

// var defaultM2Shell = [...]string{"cmd.exe", "/C", "start", "/D", "<DIR>",
//
//	"L:\\prog\\msys64\\mingw64.exe"}
var defaultLinuxShell = [...]string{"mate-terminal"}
var defaultWinEditCmd = [...]string{"notepad", "<FILE>"}
var defaultLinuxEditCmd = [...]string{"gedit", "<FILE>"}
var defaultWinBatchCmd = [...]string{"cmd.exe", "/C", "<PATH>"}
var defaultLinuxBatchCmd = [...]string{"bash", "-c", "<PATH>"}

// the browser is user selectable in prefsDialog
var defaultWinBrowser = "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe"
var defaultLinuxBrowser = "firefox"
var defaultWinBrowseCmd = [...]string{"<BROWSER>", "<URL>"}
var defaultLinuxBrowseCmd = [...]string{"bash", "-c", "<PATH>"}

func NewPrefs(path string) *Prefs {
	p := Prefs{}
	p.DateTimeFormat = defaultDateTimeFormat
	p.HiddenFiles = defaultHiddenFiles
	p.Path = path
	p.History = make([]string, 0)
	p.Favorites = make([]string, 0)
	p.Assoc = make([]FileAssoc, 0)
	p.Text = 16
	setDefaultAssociations(&p)
	setDefaultEditCmd(&p)
	setDefaultBatchCmd(&p)
	setDefaultBrowsing(&p)
	widgets(&p)
	return &p
}
func setDefaultBrowsing(p *Prefs) {
	switch runtime.GOOS {
	case "windows":
		p.Browse = make([]string, len(defaultWinBrowseCmd))
		copy(p.Browse, defaultWinBrowseCmd[0:])
		p.Browser = defaultWinBrowser
	default:
		p.Browse = make([]string, len(defaultLinuxBrowseCmd))
		copy(p.Browse, defaultLinuxBrowseCmd[0:])
		p.Browser = defaultLinuxBrowser
	}
}
func setDefaultBatchCmd(p *Prefs) {
	switch runtime.GOOS {
	case "windows":
		p.Batch = make([]string, len(defaultWinBatchCmd))
		copy(p.Batch, defaultWinBatchCmd[0:])
	default:
		p.Batch = make([]string, len(defaultLinuxBatchCmd))
		copy(p.Batch, defaultLinuxBatchCmd[0:])
	}
}
func setDefaultEditCmd(p *Prefs) {
	switch runtime.GOOS {
	case "windows":
		p.Edit = make([]string, len(defaultWinEditCmd))
		copy(p.Edit, defaultWinEditCmd[0:])
	default:
		p.Edit = make([]string, len(defaultLinuxEditCmd))
		copy(p.Edit, defaultLinuxEditCmd[0:])
	}
}
