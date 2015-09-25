# wingo
Simple wrapper for w32 package for personal use.

## Lil bit o' docs

To create a new window call the function:  
```func NewWindow(resizable, minimizable bool) *Window```  
Methods of `Window` are defined in `window.go` file.

Window menu is created with the functions:  
```func NewMenu() *Menu```  
```func (m *Menu) AppendItemText(t string) *MenuItem```  
```func (m *Menu) AppendPopup(t string) *Menu```  
Methods of `Menu` and `MenuItem` are defined in `menu.go` file.

Four available types of controls are created with functions:  
```func (w *Window) NewLabel() *Label```  
```func (w *Window) NewButton() *Button```  
```func (w *Window) NewTextEdit() *TextEdit```  
```func (w *Window) NewListBox() *ListBox```  
Methods of the control structures are defined in `controls.go` file.

After the window is set up, call the `wingo.Start()` to begin event loop.  
The application is haltable anytime with a call to `wingo.Exit()`  
These two functions are defined in `app.go`.

## Build tools

The folder `/build-tools' contains three files, inserted for fast building of an application with window resources.  
* `app.manifest` - sample manifest file.
* `resources.rc` - sample resources file, containing a reference to the manifest.
* `build.bat` - sample build script, which compiles resources file with `windres` into an object, then compiles the main.go file and links program with resources.
