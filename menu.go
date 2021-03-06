package wingo

import (
	"errors"
	"github.com/yamnikov-oleg/w32"
)

var menuItems []*MenuItem
var miCount uint = 0

type Menu struct {
	handle w32.HMENU
	Items  []*MenuItem
}

type MenuItem struct {
	Text    string
	id      uintptr
	Submenu *Menu
	OnClick func(*MenuItem)
}

func findMenuItemById(id uintptr) (*MenuItem, error) {
	for _, mi := range menuItems {
		if mi.id == id {
			return mi, nil
		}
	}
	return nil, errors.New("Could not find menuItem with such id")
}

func NewMenu() *Menu {
	m := new(Menu)
	m.handle = w32.CreateMenu()
	return m
}

func NewContextMenu() *Menu {
	return NewMenu().AppendPopup("")
}

func (m *Menu) AppendItemText(t string) *MenuItem {
	mi := new(MenuItem)
	mi.Text = t
	mi.id = uintptr(miCount)
	miCount++
	if w32.AppendMenu(m.handle, w32.MF_STRING, mi.id, t) {
		menuItems = append(menuItems, mi)
		m.Items = append(m.Items, mi)
	}
	return mi
}

func (m *Menu) AppendPopup(t string) *Menu {
	menu := NewMenu()
	mi := new(MenuItem)
	mi.Text = t
	mi.id = uintptr(menu.handle)
	mi.Submenu = menu
	w32.AppendMenu(m.handle, w32.MF_STRING|w32.MF_POPUP, mi.id, t)
	return menu
}

func (m *Menu) StartContext(wnd *Window) {
	x, y, ok := w32.GetCursorPos()
	if ok {
		wnd.Focus()
		w32.TrackPopupMenu(m.handle, 0, x, y, wnd.handle)
	}
}
