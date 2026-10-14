package imaging

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/draw"
)

func swapsAxes(orientation int) bool { return orientation >= 5 && orientation <= 8 }

// fit returns w,h scaled so the longest edge is at most maxEdge. It never
// upscales and never returns a zero edge.
func fit(w, h, maxEdge int) (int, int) {
	if w <= maxEdge && h <= maxEdge {
		return w, h
	}
	if w >= h {
		return maxEdge, max(1, int(math.Round(float64(h)*float64(maxEdge)/float64(w))))
	}
	return max(1, int(math.Round(float64(w)*float64(maxEdge)/float64(h)))), maxEdge
}

// scaleOntoWhite draws src onto an opaque white w×h canvas, compositing any
// transparency onto white in the same pass as the resize.
func scaleOntoWhite(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	sb := src.Bounds()
	if sb.Dx() == w && sb.Dy() == h {
		draw.Draw(dst, dst.Bounds(), src, sb.Min, draw.Over)
		return dst
	}
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, sb, draw.Over, nil)
	return dst
}

// orient applies an EXIF orientation transform. src must have Min = (0,0).
func orient(src *image.RGBA, o int) *image.RGBA {
	if o < 2 || o > 8 {
		return src
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dw, dh := w, h
	if swapsAxes(o) {
		dw, dh = h, w
	}
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		for x := 0; x < dw; x++ {
			var sx, sy int
			switch o {
			case 2:
				sx, sy = w-1-x, y
			case 3:
				sx, sy = w-1-x, h-1-y
			case 4:
				sx, sy = x, h-1-y
			case 5:
				sx, sy = y, x
			case 6:
				sx, sy = y, h-1-x
			case 7:
				sx, sy = w-1-y, h-1-x
			case 8:
				sx, sy = w-1-y, x
			}
			dst.SetRGBA(x, y, src.RGBAAt(sx, sy))
		}
	}
	return dst
}
