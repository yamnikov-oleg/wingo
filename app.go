package wingo

import (
	"github.com/yamnikov-oleg/w32"
	"unsafe"
)

type Vector struct {
	X, Y int
}

var appHandle w32.HINSTANCE
var (
	defaultFont w32.HFONT
	boldFont    w32.HFONT
)

func InfoMessage(title, text string) {
	w32.MessageBox(0, text, title, w32.MB_ICONINFORMATION|w32.MB_OK)
}

func Start() {
	var msg w32.MSG
	for w32.GetMessage(&msg, 0, 0, 0) != 0 {
		w32.TranslateMessage(&msg)
		w32.DispatchMessage(&msg)
	}
}

func Exit() {
	w32.PostQuitMessage(0)
}

func initCommonControls() {
	var ccstr w32.INITCOMMONCONTROLSEX
	ccstr = w32.INITCOMMONCONTROLSEX{
		/* DwSize */ uint32(unsafe.Sizeof(ccstr)),
		/* DwICC */ w32.ICC_WIN95_CLASSES,
	}
	w32.InitCommonControlsEx(&ccstr)
}

func loadDefaultFonts() {
	var spi w32.SYSTEMPARAMETERSINFO
	spi.Size = uint32(unsafe.Sizeof(spi))
	w32.SystemParametersInfo(w32.SPI_GETNONCLIENTMETRICS, spi.Size, &spi, 0)

	logFont := spi.MessageFont
	defaultFont = w32.CreateFontIndirect(&logFont)

	logFont.Weight = w32.FW_BOLD
	boldFont = w32.CreateFontIndirect(&logFont)
}

func init() {
	initCommonControls()
	loadDefaultFonts()
	appHandle = w32.GetModuleHandle("")
	if err := registerClasses(); err != nil {
		panic(err)
	}
}
