package imaging

import (
	"testing"
)

func TestJPEGOrientation(t *testing.T) {
	base := encodeJPEG(t, quadrantImage(32, 32))

	for _, be := range []bool{false, true} {
		for o := uint16(1); o <= 8; o++ {
			got := jpegOrientation(injectAPP1(base, exifAPP1(o, be)))
			if got != int(o) {
				t.Errorf("bigEndian=%v orientation=%d: got %d", be, o, got)
			}
		}
	}

	cases := []struct {
		name string
		src  []byte
	}{
		{"no APP1", base},
		{"empty", nil},
		{"not a jpeg", []byte("GIF89a....")},
		{"out of range 0", injectAPP1(base, exifAPP1(0, false))},
		{"out of range 9", injectAPP1(base, exifAPP1(9, false))},
		{"wrong type", injectAPP1(base, mutate(exifAPP1(6, false), 22, 4))},  // type LONG
		{"wrong count", injectAPP1(base, mutate(exifAPP1(6, false), 24, 2))}, // count 2
		{"bad byte order", injectAPP1(base, mutate(exifAPP1(6, false), 10, 'X'))},
		{"non-exif APP1", injectAPP1(base, mutate(exifAPP1(6, false), 4, 'X'))},
	}
	for _, c := range cases {
		if got := jpegOrientation(c.src); got != 1 {
			t.Errorf("%s: got %d, want 1", c.name, got)
		}
	}
}

func TestJPEGOrientationTruncatedAtEveryOffset(t *testing.T) {
	full := injectAPP1(encodeJPEG(t, quadrantImage(16, 16)), exifAPP1(6, false))
	for n := 0; n <= len(full); n++ {
		got := jpegOrientation(full[:n])
		if got < 1 || got > 8 {
			t.Fatalf("truncated at %d: got %d", n, got)
		}
	}
}

// mutate returns a copy of b with byte i set to v. Offsets into an
// exifAPP1(…, false) segment: 0-1 marker, 2-3 length, 4-9 "Exif\0\0",
// 10-13 TIFF byte order + magic, 14-17 IFD0 offset, 18-19 entry count,
// 20-21 tag, 22-23 type, 24-27 count, 28-29 value (little-endian, so the
// low byte of each field comes first).
func mutate(b []byte, i int, v byte) []byte {
	out := append([]byte(nil), b...)
	out[i] = v
	return out
}

func FuzzJPEGOrientation(f *testing.F) {
	f.Add(injectAPP1(encodeJPEG(f, quadrantImage(16, 16)), exifAPP1(6, false)))
	f.Add(injectAPP1(encodeJPEG(f, quadrantImage(16, 16)), exifAPP1(3, true)))
	f.Add([]byte{0xFF, 0xD8, 0xFF, 0xE1, 0x00, 0x02})
	f.Fuzz(func(t *testing.T, b []byte) {
		got := jpegOrientation(b)
		if got < 1 || got > 8 {
			t.Fatalf("got %d", got)
		}
	})
}
