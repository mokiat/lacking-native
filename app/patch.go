package app

import (
	"image/color"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// The following wrapper attempts to workaround
// https://github.com/veandco/go-sdl2/issues/580

func WrapSurface(surface *sdl.Surface) WrapperSurface {
	return WrapperSurface{
		Surface: surface,
	}
}

type WrapperSurface struct {
	*sdl.Surface
}

// Set the color of the pixel at (x, y) using this surface's color model to
// convert c to the appropriate color. This method is required for the
// draw.Image interface. The surface may require locking before calling Set.
func (s WrapperSurface) Set(x, y int, c color.Color) {
	nrgbaColor := color.NRGBAModel.Convert(c).(color.NRGBA)
	colR, colG, colB, colA := nrgbaColor.R, nrgbaColor.G, nrgbaColor.B, nrgbaColor.A

	pix := s.Pixels()
	i := int32(y)*s.Pitch + int32(x)*int32(s.Format.BytesPerPixel)
	switch s.Format.Format {
	case sdl.PIXELFORMAT_ARGB8888:
		pix[i+0] = colB
		pix[i+1] = colG
		pix[i+2] = colR
		pix[i+3] = colA
	case sdl.PIXELFORMAT_ABGR8888:
		pix[i+3] = colR
		pix[i+2] = colG
		pix[i+1] = colB
		pix[i+0] = colA
	case sdl.PIXELFORMAT_RGB24, sdl.PIXELFORMAT_RGB888:
		pix[i+0] = colB
		pix[i+1] = colG
		pix[i+2] = colR
	case sdl.PIXELFORMAT_BGR24, sdl.PIXELFORMAT_BGR888:
		pix[i+2] = colR
		pix[i+1] = colG
		pix[i+0] = colB
	case sdl.PIXELFORMAT_RGB444:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		r := uint32(colR) >> 4 & 0x0F
		g := uint32(colG) >> 4 & 0x0F
		b := uint32(colB) >> 4 & 0x0F
		*buf = r<<8 | g<<4 | b
	case sdl.PIXELFORMAT_RGB332:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		r := uint32(colR) >> 5 & 0x0F
		g := uint32(colG) >> 5 & 0x0F
		b := uint32(colB) >> 6 & 0x0F
		*buf = r<<5 | g<<2 | b
	case sdl.PIXELFORMAT_RGB565:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		r := uint32(colR) >> 3 & 0xFF
		g := uint32(colG) >> 2 & 0xFF
		b := uint32(colB) >> 3 & 0xFF
		*buf = r<<11 | g<<5 | b
	case sdl.PIXELFORMAT_RGB555:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		r := uint32(colR) >> 3 & 0xFF
		g := uint32(colG) >> 3 & 0xFF
		b := uint32(colB) >> 3 & 0xFF
		*buf = r<<10 | g<<5 | b
	case sdl.PIXELFORMAT_BGR565:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		r := uint32(colR) >> 3 & 0xFF
		g := uint32(colG) >> 2 & 0xFF
		b := uint32(colB) >> 3 & 0xFF
		*buf = b<<11 | g<<5 | r
	case sdl.PIXELFORMAT_BGR555:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		r := uint32(colR) >> 3 & 0xFF
		g := uint32(colG) >> 3 & 0xFF
		b := uint32(colB) >> 3 & 0xFF
		*buf = b<<10 | g<<5 | r
	case sdl.PIXELFORMAT_ARGB4444:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		a := uint32(colA) >> 4 & 0x0F
		r := uint32(colR) >> 4 & 0x0F
		g := uint32(colG) >> 4 & 0x0F
		b := uint32(colB) >> 4 & 0x0F
		*buf = a<<12 | r<<8 | g<<4 | b
	case sdl.PIXELFORMAT_ABGR4444:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		a := uint32(colA) >> 4 & 0x0F
		r := uint32(colR) >> 4 & 0x0F
		g := uint32(colG) >> 4 & 0x0F
		b := uint32(colB) >> 4 & 0x0F
		*buf = a<<12 | b<<8 | g<<4 | r
	case sdl.PIXELFORMAT_RGBA4444:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		r := uint32(colR) >> 4 & 0x0F
		g := uint32(colG) >> 4 & 0x0F
		b := uint32(colB) >> 4 & 0x0F
		a := uint32(colA) >> 4 & 0x0F
		*buf = r<<12 | g<<8 | b<<4 | a
	case sdl.PIXELFORMAT_BGRA4444:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		r := uint32(colR) >> 4 & 0x0F
		g := uint32(colG) >> 4 & 0x0F
		b := uint32(colB) >> 4 & 0x0F
		a := uint32(colA) >> 4 & 0x0F
		*buf = b<<12 | g<<8 | r<<4 | a
	case sdl.PIXELFORMAT_ARGB1555:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		r := uint32(colR) >> 3 & 0xFF
		g := uint32(colG) >> 3 & 0xFF
		b := uint32(colB) >> 3 & 0xFF
		a := uint32(0)
		if colA > 0 {
			a = 1
		}
		*buf = a<<15 | r<<10 | g<<5 | b
	case sdl.PIXELFORMAT_RGBA5551:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		r := uint32(colR) >> 3 & 0xFF
		g := uint32(colG) >> 3 & 0xFF
		b := uint32(colB) >> 3 & 0xFF
		a := uint32(0)
		if colA > 0 {
			a = 1
		}
		*buf = r<<11 | g<<6 | b<<1 | a
	case sdl.PIXELFORMAT_ABGR1555:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		r := uint32(colR) >> 3 & 0xFF
		g := uint32(colG) >> 3 & 0xFF
		b := uint32(colB) >> 3 & 0xFF
		a := uint32(0)
		if colA > 0 {
			a = 1
		}
		*buf = a<<15 | b<<10 | g<<5 | r
	case sdl.PIXELFORMAT_BGRA5551:
		buf := (*uint32)(unsafe.Pointer(&pix[i]))
		r := uint32(colR) >> 3 & 0xFF
		g := uint32(colG) >> 3 & 0xFF
		b := uint32(colB) >> 3 & 0xFF
		a := uint32(0)
		if colA > 0 {
			a = 1
		}
		*buf = b<<11 | g<<6 | r<<1 | a
	case sdl.PIXELFORMAT_RGBA8888:
		pix[i+3] = colR
		pix[i+2] = colG
		pix[i+1] = colB
		pix[i+0] = colA
	case sdl.PIXELFORMAT_BGRA8888:
		pix[i+3] = colB
		pix[i+2] = colG
		pix[i+1] = colR
		pix[i+0] = colA
	default:
		panic("Unknown pixel format!")
	}
}
