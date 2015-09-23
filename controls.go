package wingo

import (
	"github.com/AllenDang/w32"
	"syscall"
	"unsafe"
)

var (
	labelClassName, _    = syscall.UTF16PtrFromString("STATIC")
	buttonClassName, _   = syscall.UTF16PtrFromString("BUTTON")
	textEditClassName, _ = syscall.UTF16PtrFromString("EDIT")
	listBoxClassName, _  = syscall.UTF16PtrFromString("LISTBOX")
)

var controlCount uint = 0

func getNewControlId() uintptr {
	controlCount++
	return uintptr(controlCount)
}

type controlI interface {
	dispatchEvent(handle w32.HWND, notifCode int)
}

var controls []controlI

func dispatchControlEvent(handle w32.HWND, notifCode int) {
	for _, c := range controls {
		c.dispatchEvent(handle, notifCode)
	}
}

type control struct {
	handle w32.HWND
	id     uintptr
	Parent *Window
}

func (c *control) GetText() string {
	return w32.GetWindowText(c.handle)
}

func (c *control) SetText(s string) {
	w32.SetWindowText(c.handle, s)
}

func (c *control) GetPosition() Vector {
	rect := w32.GetWindowRect(c.handle)
	return Vector{int(rect.Left), int(rect.Top)}
}

func (c *control) SetPosition(p Vector) {
	w32.SetWindowPos(c.handle, 0, p.X, p.Y, 0, 0, w32.SWP_NOSIZE|w32.SWP_NOZORDER)
}

func (c *control) GetSize() Vector {
	rect := w32.GetWindowRect(c.handle)
	return Vector{int(rect.Right - rect.Left), int(rect.Bottom - rect.Top)}
}

func (c *control) SetSize(s Vector) {
	w32.SetWindowPos(c.handle, 0, 0, 0, s.X, s.Y, w32.SWP_NOMOVE|w32.SWP_NOZORDER)
}

func (c *control) setFont(f w32.HFONT) {
	w32.SendMessage(c.handle, w32.WM_SETFONT, uintptr(f), w32.MAKELONG(1, 0))
}

func (c *control) SetBold(b bool) {
	if b {
		c.setFont(boldFont)
	} else {
		c.setFont(defaultFont)
	}
}

type Label struct {
	control
}

func (l *Label) dispatchEvent(handle w32.HWND, notifCode int) {
	return
}

type Button struct {
	control
	OnClick func(*Button)
}

func (b *Button) dispatchEvent(handle w32.HWND, notifCode int) {
	if handle != b.handle {
		return
	}
	switch notifCode {
	case w32.BN_CLICKED:
		if b.OnClick != nil {
			b.OnClick(b)
		}
	}
}

type TextEdit struct {
	control
	OnChange func(*TextEdit, string)
}

func (te *TextEdit) dispatchEvent(handle w32.HWND, notifCode int) {
	if handle != te.handle {
		return
	}
	switch notifCode {
	case w32.EN_CHANGE:
		if te.OnChange != nil {
			te.OnChange(te, te.GetText())
		}
	}
}

type ListBox struct {
	control
	OnSelected func(*ListBox, string, int)
}

func (lb *ListBox) dispatchEvent(handle w32.HWND, notifCode int) {
	if handle != lb.handle {
		return
	}
	switch notifCode {
	case w32.LBN_SELCHANGE:
		if lb.OnSelected != nil {
			t, i := lb.GetSelection()
			lb.OnSelected(lb, t, i)
		}
	}
}

func (lb *ListBox) MakeDraggable() {
	w32.MakeDragList(lb.handle)
}

func (lb *ListBox) Clear() {
	w32.SendMessage(lb.handle, w32.LB_RESETCONTENT, 0, 0)
}

func (lb *ListBox) Get(i int) (text string, selected bool) {
	bufSize := 256
	buffer := make([]uint16, bufSize)
	w32.SendMessage(lb.handle, w32.LB_GETTEXT, uintptr(i), uintptr(unsafe.Pointer(&buffer[0])))

	text = syscall.UTF16ToString(buffer)
	selected = w32.SendMessage(lb.handle, w32.LB_GETSEL, uintptr(i), 0) != 0

	return
}

func (lb *ListBox) Set(i int, t string) {
	items := lb.GetList()
	_, sel := lb.GetSelection()

	if i >= 0 && i < len(items) {
		items[i] = t
	}

	lb.SetList(items)
	lb.SetSelection(sel)
}

func (lb *ListBox) Append(t string) int {
	tptr, _ := syscall.UTF16PtrFromString(t)
	w32.SendMessage(lb.handle, w32.LB_ADDSTRING, 0, uintptr(unsafe.Pointer(tptr)))
	return lb.Count() - 1
}

func (lb *ListBox) Delete(i int) {
	_, sel := lb.GetSelection()

	w32.SendMessage(lb.handle, w32.LB_DELETESTRING, uintptr(i), 0)

	switch {
	case sel == -1:
		return
	case sel == i:
		lb.SetSelection(-1)
	case sel < i:
		lb.SetSelection(sel)
	case sel > i:
		lb.SetSelection(sel - 1)
	}
}

func (lb *ListBox) GetSelection() (text string, i int) {
	count := lb.Count()
	var s bool
	for i = 0; i < count; i++ {
		text, s = lb.Get(i)
		if s {
			return
		}
	}
	return "", -1
}

func (lb *ListBox) SetSelection(i int) {
	w32.SendMessage(lb.handle, w32.LB_SETCURSEL, uintptr(i), 0)
}

func (lb *ListBox) Count() int {
	return int(w32.SendMessage(lb.handle, w32.LB_GETCOUNT, 0, 0))
}

func (lb *ListBox) GetList() []string {
	count := lb.Count()
	list := make([]string, count)
	for i, _ := range list {
		list[i], _ = lb.Get(i)
	}
	return list
}

func (lb *ListBox) SetList(items []string) {
	lb.Clear()

	for _, it := range items {
		lb.Append(it)
	}
}
