package wingo

import (
	"errors"
	"github.com/yamnikov-oleg/w32"
	"syscall"
	"unsafe"
)

const WM_TRAYICON = w32.WM_USER + 1

var wndClassName, _ = syscall.UTF16PtrFromString("WingoWindow")
var windows []*Window

type Window struct {
	handle      w32.HWND
	resizable   bool
	minimizable bool
	controls    []controlI

	OnSizeChanged func(*Window, Vector)
	OnMinimize    func(*Window)
	OnMaximize    func(*Window)
	OnTrayClick   func(*Window)
	OnClose       func(*Window) bool
	OnDestroy     func(*Window)
}

func registerClasses() error {
	var wc w32.WNDCLASSEX
	wc = w32.WNDCLASSEX{
		/* Size */ uint32(unsafe.Sizeof(wc)),
		/* Style */ w32.CS_HREDRAW | w32.CS_VREDRAW,
		/* WndProc */ syscall.NewCallback(wndProc),
		/* ClsExtra */ 0,
		/* WndExtra */ 0,
		/* Instance */ appHandle,
		/* Icon */ 0,
		/* Cursor */ w32.LoadCursor(0, w32.MakeIntResource(w32.IDC_ARROW)),
		/* Background*/ w32.HBRUSH(w32.COLOR_WINDOW),
		/* MenuName */ nil,
		/* ClassName */ wndClassName,
		/* IconSm */ 0,
	}

	if res := w32.RegisterClassEx(&wc); res == 0 {
		return errors.New("Failed to register window class!")
	}

	return nil
}

func findWndByHandle(h w32.HWND) (*Window, error) {
	for _, w := range windows {
		if w.handle == h {
			return w, nil
		}
	}
	return nil, errors.New("No window with such handle")
}

func wndProc(hwnd w32.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	wnd, err := findWndByHandle(hwnd)
	if err != nil {
		return w32.DefWindowProc(hwnd, msg, wParam, lParam)
	}

	switch msg {
	case w32.WM_SIZE:
		if wParam == w32.SIZE_MINIMIZED {
			if wnd.OnMinimize != nil {
				wnd.OnMinimize(wnd)
			}
			// We don't want to send OnSizeChanged notification then
			return 0
		} else if wParam == w32.SIZE_MAXIMIZED {
			if wnd.OnMaximize != nil {
				wnd.OnMaximize(wnd)
			}
		}
		w := int(w32.LOWORD(uint32(lParam)))
		h := int(w32.HIWORD(uint32(lParam)))
		if wnd.OnSizeChanged != nil {
			wnd.OnSizeChanged(wnd, Vector{w, h})
		}
		return 0
	case w32.WM_COMMAND:
		notifCode := int(w32.HIWORD(uint32(wParam)))
		controlId := uintptr(w32.LOWORD(uint32(wParam)))
		controlHandle := w32.HWND(lParam)

		if controlHandle == 0 {
			if notifCode != 0 {
				return 0
			}
			// Menu item click
			mi, err := findMenuItemById(controlId)
			if err == nil && mi.OnClick != nil {
				mi.OnClick(mi)
			}
		} else {
			// Control notification
			dispatchControlEvent(controlHandle, notifCode)
		}

		return 0
	case WM_TRAYICON:
		switch w32.LOWORD(uint32(lParam)) {
		case w32.WM_LBUTTONUP:
			if wnd.OnTrayClick != nil {
				wnd.OnTrayClick(wnd)
			}
		}
		return 0
	case w32.WM_CLOSE:
		if wnd.OnClose == nil || wnd.OnClose(wnd) {
			wnd.Destroy()
		}
		return 0
	case w32.WM_DESTROY:
		if wnd.OnDestroy != nil {
			wnd.OnDestroy(wnd)
		} else {
			Exit()
		}
		return 0
	default:
		return w32.DefWindowProc(hwnd, msg, wParam, lParam)
	}
}

func (w *Window) compileStyle() (style uint) {
	style = w32.WS_OVERLAPPED | w32.WS_CAPTION | w32.WS_SYSMENU
	if w.resizable {
		style |= w32.WS_SIZEBOX | w32.WS_MAXIMIZEBOX
	}
	if w.minimizable {
		style |= w32.WS_MINIMIZEBOX
	}
	return
}

func NewWindow(rsz, min bool) *Window {
	wnd := new(Window)
	wnd.resizable = rsz
	wnd.minimizable = min
	style := wnd.compileStyle()

	wnd.handle = w32.CreateWindowEx(
		/* exStyle */ 0,
		/* className */ wndClassName,
		/* windowName */ nil,
		/* style */ style,
		/* x */ w32.CW_USEDEFAULT,
		/* y */ w32.CW_USEDEFAULT,
		/* width */ w32.CW_USEDEFAULT,
		/* height */ w32.CW_USEDEFAULT,
		/* parent */ 0,
		/* menu */ 0,
		/* instance */ appHandle,
		/* param */ nil,
	)

	if wnd.handle == 0 {
		panic("Can't create window!")
	}

	windows = append(windows, wnd)

	return wnd
}

func (w *Window) Destroy() {
	w32.DestroyWindow(w.handle)
}

func (w *Window) Show() {
	w32.ShowWindow(w.handle, w32.SW_SHOW)
}

func (w *Window) Hide() {
	w32.ShowWindow(w.handle, w32.SW_HIDE)
}

func (w *Window) Restore() {
	w32.ShowWindow(w.handle, w32.SW_RESTORE)
}

func (w *Window) GetTitle() string {
	return w32.GetWindowText(w.handle)
}

func (w *Window) SetTitle(s string) {
	w32.SetWindowText(w.handle, s)
}

func (w *Window) GetPosition() Vector {
	rect := w32.GetWindowRect(w.handle)
	return Vector{int(rect.Left), int(rect.Top)}
}

func (w *Window) SetPosition(p Vector) {
	w32.SetWindowPos(w.handle, 0, p.X, p.Y, 0, 0, w32.SWP_NOSIZE|w32.SWP_NOZORDER)
}

func (w *Window) GetSize() Vector {
	rect := w32.GetWindowRect(w.handle)
	return Vector{int(rect.Right - rect.Left), int(rect.Bottom - rect.Top)}
}

func (w *Window) SetSize(s Vector) {
	w32.SetWindowPos(w.handle, 0, 0, 0, s.X, s.Y, w32.SWP_NOMOVE|w32.SWP_NOZORDER)
}

func (w *Window) GetClientSize() Vector {
	rect := w32.GetClientRect(w.handle)
	return Vector{int(rect.Right - rect.Left), int(rect.Bottom - rect.Top)}
}

func (w *Window) ApplyMenu(m *Menu) {
	w32.SetMenu(w.handle, m.handle)
}

func (w *Window) SetIcon(icon Icon) {
	w32.SendMessage(w.handle, w32.WM_SETICON, w32.ICON_BIG, uintptr(icon))
	w32.SendMessage(w.handle, w32.WM_SETICON, w32.ICON_SMALL, uintptr(icon))
	w32.SendMessage(w.handle, w32.WM_SETICON, w32.ICON_SMALL2, uintptr(icon))
}

func (w *Window) AddTrayIcon(ico Icon, tip string) {
	var nid w32.NOTIFYICONDATA
	nid = w32.NOTIFYICONDATA{
		/* Size */ uint32(unsafe.Sizeof(nid)),
		/* Wnd */ w.handle,
		/* ID */ 1,
		/* Flags */ w32.NIF_MESSAGE | w32.NIF_ICON | w32.NIF_TIP | w32.NIF_SHOWTIP,
		/* CallbackMessage */ WM_TRAYICON,
		/* Icon */ w32.HICON(ico),
		/* Tip */ [128]uint16{},
		/* State */ 0,
		/* StateMask */ 0,
		/* Info */ [256]uint16{},
		/* TimeoutOrVersion */ w32.NOTIFYICON_VERSION_4,
		/* InfoTitle */ [64]uint16{},
		/* InfoFlags */ 0,
		/* Item */ w32.GUID{},
		/* BalloonIcon */ 0,
	}

	tipbuf := syscall.StringToUTF16(tip)
	copy(nid.Tip[:], tipbuf)

	w32.Shell_NotifyIcon(w32.NIM_ADD, &nid)
}

func (w *Window) NewLabel() *Label {
	l := new(Label)
	l.id = getNewControlId()
	l.handle = w32.CreateWindowEx(
		/* exStyle */ 0,
		/* className */ labelClassName,
		/* windowName */ nil,
		/* style */ w32.WS_CHILD|w32.WS_VISIBLE,
		/* x */ w32.CW_USEDEFAULT,
		/* y */ w32.CW_USEDEFAULT,
		/* width */ w32.CW_USEDEFAULT,
		/* height */ w32.CW_USEDEFAULT,
		/* parent */ w.handle,
		/* menu */ w32.HMENU(l.id),
		/* instance */ appHandle,
		/* param */ nil,
	)
	controls = append(controls, l)
	l.Parent = w
	l.setFont(defaultFont)
	return l
}

func (w *Window) NewButton() *Button {
	b := new(Button)
	b.id = getNewControlId()
	b.handle = w32.CreateWindowEx(
		/* exStyle */ 0,
		/* className */ buttonClassName,
		/* windowName */ nil,
		/* style */ w32.WS_CHILD|w32.WS_VISIBLE|w32.BS_PUSHBUTTON,
		/* x */ w32.CW_USEDEFAULT,
		/* y */ w32.CW_USEDEFAULT,
		/* width */ w32.CW_USEDEFAULT,
		/* height */ w32.CW_USEDEFAULT,
		/* parent */ w.handle,
		/* menu */ w32.HMENU(b.id),
		/* instance */ appHandle,
		/* param */ nil,
	)
	controls = append(controls, b)
	b.Parent = w
	b.setFont(defaultFont)
	return b
}

func (w *Window) NewTextEdit() *TextEdit {
	te := new(TextEdit)
	te.id = getNewControlId()
	te.handle = w32.CreateWindowEx(
		/* exStyle */ w32.WS_EX_CLIENTEDGE,
		/* className */ textEditClassName,
		/* windowName */ nil,
		/* style */ w32.WS_CHILD|w32.WS_VISIBLE|w32.ES_AUTOHSCROLL,
		/* x */ w32.CW_USEDEFAULT,
		/* y */ w32.CW_USEDEFAULT,
		/* width */ w32.CW_USEDEFAULT,
		/* height */ w32.CW_USEDEFAULT,
		/* parent */ w.handle,
		/* menu */ w32.HMENU(te.id),
		/* instance */ appHandle,
		/* param */ nil,
	)
	controls = append(controls, te)
	te.Parent = w
	te.setFont(defaultFont)
	return te
}

func (w *Window) NewListBox() *ListBox {
	lb := new(ListBox)
	lb.id = getNewControlId()
	lb.handle = w32.CreateWindowEx(
		/* exStyle */ w32.WS_EX_CLIENTEDGE,
		/* className */ listBoxClassName,
		/* windowName */ nil,
		/* style */ w32.WS_CHILD|w32.WS_VISIBLE|w32.WS_VSCROLL|w32.LBS_NOTIFY,
		/* x */ w32.CW_USEDEFAULT,
		/* y */ w32.CW_USEDEFAULT,
		/* width */ w32.CW_USEDEFAULT,
		/* height */ w32.CW_USEDEFAULT,
		/* parent */ w.handle,
		/* menu */ w32.HMENU(lb.id),
		/* instance */ appHandle,
		/* param */ nil,
	)
	controls = append(controls, lb)
	lb.Parent = w
	lb.setFont(defaultFont)
	return lb
}
