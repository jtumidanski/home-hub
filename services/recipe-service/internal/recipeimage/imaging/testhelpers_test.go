package imaging

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

// exifAPP1 builds a complete APP1 segment (marker, length, payload) carrying
// a single IFD0 Orientation entry.
func exifAPP1(orientation uint16, bigEndian bool) []byte {
	var bo binary.AppendByteOrder = binary.LittleEndian
	tiff := []byte("II*\x00")
	if bigEndian {
		bo = binary.BigEndian
		tiff = []byte("MM\x00*")
	}
	tiff = bo.AppendUint32(tiff, 8)      // IFD0 offset
	tiff = bo.AppendUint16(tiff, 1)      // entry count
	tiff = bo.AppendUint16(tiff, 0x0112) // Orientation tag
	tiff = bo.AppendUint16(tiff, 3)      // SHORT
	tiff = bo.AppendUint32(tiff, 1)      // count
	tiff = bo.AppendUint16(tiff, orientation)
	tiff = append(tiff, 0, 0)       // value padding to 4 bytes
	tiff = bo.AppendUint32(tiff, 0) // next IFD
	payload := append([]byte("Exif\x00\x00"), tiff...)
	seg := []byte{0xFF, 0xE1}
	seg = binary.BigEndian.AppendUint16(seg, uint16(len(payload)+2))
	return append(seg, payload...)
}

// injectAPP1 inserts a segment right after SOI.
func injectAPP1(jpg, app1 []byte) []byte {
	out := make([]byte, 0, len(jpg)+len(app1))
	out = append(out, jpg[:2]...)
	out = append(out, app1...)
	return append(out, jpg[2:]...)
}

func encodeJPEG(t testing.TB, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// quadrantImage is blue with a solid red block in the top-left quarter, so a
// test can tell where the original top-left corner ended up.
func quadrantImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := color.RGBA{0, 0, 255, 255}
			if x < w/2 && y < h/2 {
				c = color.RGBA{255, 0, 0, 255}
			}
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

// pngHeaderOnly returns a PNG signature plus a valid IHDR chunk for an 8-bit
// RGB image of the given size, with no pixel data. DecodeConfig accepts it;
// a full decode would fail.
func pngHeaderOnly(w, h uint32) []byte {
	b := []byte("\x89PNG\r\n\x1a\n")
	ihdr := binary.BigEndian.AppendUint32(nil, w)
	ihdr = binary.BigEndian.AppendUint32(ihdr, h)
	ihdr = append(ihdr, 8, 2, 0, 0, 0)
	b = binary.BigEndian.AppendUint32(b, uint32(len(ihdr)))
	chunk := append([]byte("IHDR"), ihdr...)
	b = append(b, chunk...)
	return binary.BigEndian.AppendUint32(b, crc32.ChecksumIEEE(chunk))
}
