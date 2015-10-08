package wingo

import (
	"github.com/yamnikov-oleg/w32"
)

type Icon w32.HICON

func LoadIcon(id uint16) Icon {
	return Icon(w32.LoadIcon(appHandle, w32.MakeIntResource(id)))
}
