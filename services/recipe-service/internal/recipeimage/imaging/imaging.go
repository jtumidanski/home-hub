// Package imaging turns untrusted upload bytes into two JPEG renditions. It
// does no I/O. Decoders are registered here and nowhere else in the service.
package imaging

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

const (
	MaxPixels    = 40_000_000
	MinEdge      = 16
	FullMaxEdge  = 1600
	ThumbMaxEdge = 480
	JPEGQuality  = 82
)

var (
	ErrUnsupportedFormat = errors.New("unsupported image format")
	ErrTooLarge          = errors.New("image dimensions exceed limit")
	ErrTooSmall          = errors.New("image dimensions below minimum")
)

type Rendition struct {
	Bytes         []byte
	Width, Height int
}

type Result struct {
	Full, Thumb Rendition
}

// Process validates and converts src. Format comes only from the bytes.
// Dimension limits are enforced from the header before the full decode.
func Process(src []byte) (Result, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(src))
	if err != nil {
		return Result{}, ErrUnsupportedFormat
	}
	switch format {
	case "jpeg", "png", "webp":
	default:
		return Result{}, ErrUnsupportedFormat
	}
	if int64(cfg.Width)*int64(cfg.Height) > MaxPixels {
		return Result{}, ErrTooLarge
	}
	if min(cfg.Width, cfg.Height) < MinEdge {
		return Result{}, ErrTooSmall
	}

	img, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return Result{}, ErrUnsupportedFormat
	}

	orientation := 1
	if format == "jpeg" {
		orientation = jpegOrientation(src)
	}

	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	ow, oh := w, h
	if swapsAxes(orientation) {
		ow, oh = h, w
	}
	tw, th := fit(ow, oh, FullMaxEdge)
	sw, sh := tw, th
	if swapsAxes(orientation) {
		sw, sh = th, tw
	}

	// Resize first, then orient: rotating 1600 px is far cheaper than 4000 px,
	// and the two commute once the target axes are swapped.
	full := orient(scaleOntoWhite(img, sw, sh), orientation)

	thw, thh := fit(full.Bounds().Dx(), full.Bounds().Dy(), ThumbMaxEdge)
	thumb := scaleOntoWhite(full, thw, thh)

	fullBytes, err := encode(full)
	if err != nil {
		return Result{}, err
	}
	thumbBytes, err := encode(thumb)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Full:  Rendition{Bytes: fullBytes, Width: full.Bounds().Dx(), Height: full.Bounds().Dy()},
		Thumb: Rendition{Bytes: thumbBytes, Width: thumb.Bounds().Dx(), Height: thumb.Bounds().Dy()},
	}, nil
}

// encode re-encodes as baseline JPEG; the stdlib encoder writes no EXIF, XMP
// or ICC segments, which strips all metadata including GPS.
func encode(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: JPEGQuality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
