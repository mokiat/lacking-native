package app

import "github.com/veandco/go-sdl2/sdl"

type customCursor struct {
	surface *sdl.Surface
	cursor  *sdl.Cursor
}

func (c *customCursor) Destroy() {
	sdl.FreeCursor(c.cursor)
	c.surface.Free()
}
