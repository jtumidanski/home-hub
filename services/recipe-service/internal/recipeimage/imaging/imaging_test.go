package imaging

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"
)

// 32x24 lossless WebP: left half fully transparent, right half opaque blue.
const webpFixture = "UklGRiQAAABXRUJQVlA4TBgAAAAvH8AFEA8QMf/zHwyBQJK/9AIDRPQ/qDc="

func decodeOut(t *testing.T, b []byte) image.Image {
	t.Helper()
	img, format, err := image.Decode(bytes.NewReader(b))
	if err != nil || format != "jpeg" {
		t.Fatalf("output not a JPEG: format=%q err=%v", format, err)
	}
	return img
}

func isRed(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	return r>>8 > 180 && g>>8 < 90 && b>>8 < 90
}

func isWhite(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	return r>>8 > 235 && g>>8 > 235 && b>>8 > 235
}

func hasAPP1(b []byte) bool {
	return bytes.Contains(b, []byte{0xFF, 0xE1})
}

func TestProcessOrientation(t *testing.T) {
	base := encodeJPEG(t, quadrantImage(64, 32))
	cases := []struct {
		o          uint16
		w, h       int
		redX, redY int // a point inside where the red quadrant must land
	}{
		{1, 64, 32, 6, 6},
		{3, 64, 32, 64 - 7, 32 - 7},
		{6, 32, 64, 32 - 7, 6},
		{8, 32, 64, 6, 64 - 7},
	}
	for _, c := range cases {
		res, err := Process(injectAPP1(base, exifAPP1(c.o, false)))
		if err != nil {
			t.Fatalf("o=%d: %v", c.o, err)
		}
		if res.Full.Width != c.w || res.Full.Height != c.h {
			t.Fatalf("o=%d: full %dx%d, want %dx%d", c.o, res.Full.Width, res.Full.Height, c.w, c.h)
		}
		img := decodeOut(t, res.Full.Bytes)
		if !isRed(img.At(c.redX, c.redY)) {
			t.Errorf("o=%d: pixel (%d,%d) = %v, want red", c.o, c.redX, c.redY, img.At(c.redX, c.redY))
		}
		if hasAPP1(res.Full.Bytes) || hasAPP1(res.Thumb.Bytes) {
			t.Errorf("o=%d: output carries an APP1 (EXIF) segment", c.o)
		}
	}
}

func TestProcessOrientedDownscale(t *testing.T) {
	src := injectAPP1(encodeJPEG(t, quadrantImage(2000, 1000)), exifAPP1(6, false))
	res, err := Process(src)
	if err != nil {
		t.Fatal(err)
	}
	if res.Full.Width != 800 || res.Full.Height != 1600 {
		t.Fatalf("full %dx%d, want 800x1600", res.Full.Width, res.Full.Height)
	}
	if res.Thumb.Width != 240 || res.Thumb.Height != 480 {
		t.Fatalf("thumb %dx%d, want 240x480", res.Thumb.Width, res.Thumb.Height)
	}
	img := decodeOut(t, res.Thumb.Bytes)
	if b := img.Bounds(); b.Dx() != 240 || b.Dy() != 480 {
		t.Fatalf("decoded thumb %v", b)
	}
}

func TestProcessNeverUpscales(t *testing.T) {
	res, err := Process(encodeJPEG(t, quadrantImage(100, 50)))
	if err != nil {
		t.Fatal(err)
	}
	if res.Full.Width != 100 || res.Full.Height != 50 || res.Thumb.Width != 100 || res.Thumb.Height != 50 {
		t.Fatalf("full %dx%d thumb %dx%d", res.Full.Width, res.Full.Height, res.Thumb.Width, res.Thumb.Height)
	}
}

func TestProcessPNGAlphaCompositesWhite(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			if x >= 20 {
				img.SetNRGBA(x, y, color.NRGBA{255, 0, 0, 255})
			} // left half stays fully transparent
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	res, err := Process(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	out := decodeOut(t, res.Full.Bytes)
	if !isWhite(out.At(5, 10)) {
		t.Errorf("transparent area = %v, want white", out.At(5, 10))
	}
	if !isRed(out.At(35, 10)) {
		t.Errorf("opaque area = %v, want red", out.At(35, 10))
	}
}

func TestProcessPalettedPNGWithTransparency(t *testing.T) {
	pal := color.Palette{color.NRGBA{0, 0, 0, 0}, color.NRGBA{255, 0, 0, 255}}
	img := image.NewPaletted(image.Rect(0, 0, 32, 32), pal)
	for y := 0; y < 32; y++ {
		for x := 16; x < 32; x++ {
			img.SetColorIndex(x, y, 1)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	res, err := Process(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	out := decodeOut(t, res.Full.Bytes)
	if !isWhite(out.At(4, 16)) || !isRed(out.At(28, 16)) {
		t.Errorf("paletted composite wrong: left=%v right=%v", out.At(4, 16), out.At(28, 16))
	}

	gray := image.NewGray(image.Rect(0, 0, 20, 20))
	buf.Reset()
	if err := png.Encode(&buf, gray); err != nil {
		t.Fatal(err)
	}
	if _, err := Process(buf.Bytes()); err != nil {
		t.Fatalf("grayscale PNG: %v", err)
	}
}

func TestProcessWebP(t *testing.T) {
	src, _ := base64.StdEncoding.DecodeString(webpFixture)
	res, err := Process(src)
	if err != nil {
		t.Fatal(err)
	}
	if res.Full.Width != 32 || res.Full.Height != 24 {
		t.Fatalf("full %dx%d", res.Full.Width, res.Full.Height)
	}
	out := decodeOut(t, res.Full.Bytes)
	if !isWhite(out.At(4, 12)) {
		t.Errorf("transparent webp area = %v, want white", out.At(4, 12))
	}
}

func TestProcessRejections(t *testing.T) {
	var gifBuf bytes.Buffer
	if err := gif.Encode(&gifBuf, image.NewPaletted(image.Rect(0, 0, 32, 32), color.Palette{color.Black, color.White}), nil); err != nil {
		t.Fatal(err)
	}
	tiny := encodeJPEG(t, image.NewRGBA(image.Rect(0, 0, 8, 8)))
	bomb := pngHeaderOnly(10000, 5000)

	cases := []struct {
		name string
		src  []byte
		want error
	}{
		{"gif", gifBuf.Bytes(), ErrUnsupportedFormat},
		{"garbage", []byte("definitely not an image"), ErrUnsupportedFormat},
		{"empty", nil, ErrUnsupportedFormat},
		{"too small", tiny, ErrTooSmall},
		{"too many pixels (header only)", bomb, ErrTooLarge},
	}
	for _, c := range cases {
		if _, err := Process(c.src); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
	if len(bomb) > 64 {
		t.Fatalf("bomb fixture is %d bytes; it must be header-only to prove no full decode", len(bomb))
	}
}

func TestProcessOutputDecodesAsJPEG(t *testing.T) {
	res, err := Process(encodeJPEG(t, quadrantImage(64, 64)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := jpeg.DecodeConfig(bytes.NewReader(res.Thumb.Bytes)); err != nil {
		t.Fatal(err)
	}
}

func FuzzProcess(f *testing.F) {
	f.Add(encodeJPEG(f, quadrantImage(32, 32)))
	src, _ := base64.StdEncoding.DecodeString(webpFixture)
	f.Add(src)
	f.Add(pngHeaderOnly(100, 100))
	f.Fuzz(func(t *testing.T, b []byte) {
		res, err := Process(b)
		if err != nil {
			return
		}
		if res.Full.Width < 1 || res.Full.Width > FullMaxEdge || res.Full.Height > FullMaxEdge {
			t.Fatalf("full %dx%d", res.Full.Width, res.Full.Height)
		}
	})
}

// BenchmarkProcess12MP measures the PRD budget (< 2 s for a 12 MP JPEG).
// Record the result in services/recipe-service/docs/storage.md (Task 13).
func BenchmarkProcess12MP(b *testing.B) {
	src := injectAPP1(encodeJPEG(b, quadrantImage(4000, 3000)), exifAPP1(6, false))
	b.SetBytes(int64(len(src)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Process(src); err != nil {
			b.Fatal(err)
		}
	}
}
