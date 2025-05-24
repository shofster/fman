# FMAN
GO/fyne file manager.

FMAN is a 2 panel file manager written in the GO language and using
the fyne.io graphical toolkit.
It is designed in the manner of Midnight Commander and Total Commander.


Features include:
- Two panels, each with a separate local directory path. 
- File and Directory listings with Linux style information.
- Directory and File create, delete, and copy.
- Archive containers (zip, tar, and gzip).
- Favorite places (including User Home and known system drives / paths).
- History of recently visited places.
- Single click file selection.
- File view / edit / properties (right click).
- Double click action execution (file type dependent).
- Copy file(s) from panel to panel (no tabs)
- Display options (hidden, sort by name or date, order ascending or descending).
- Variable font size.
- Command line execution (shell started in current path).
- Preference settings for managing Favorite Places, Hidden Files, and the system path to default browser.
- Slide show of .jpeg and .png files in the current path (double click action).


Panel Controls:
- Refresh panel view.
- Select all files / directories in panel path.
- Create a new file, folder, archive in the panel path.
- Go to user's Home.
- Places menu for favorites.
- Menu of recent paths visited by that panel.
- Copy selected files to the path of the other panel.
- Delete the selected files.
- A >> or << button to set the path of the other panel to the current.
