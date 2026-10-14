package imaging

import "encoding/binary"

const tagOrientation = 0x0112

// jpegOrientation returns the EXIF Orientation (1..8) of a JPEG, or 1 when it
// is absent or anything about the structure is malformed. It examines only
// the first APP1 segment whose payload starts with "Exif\0\0", and stops at
// SOS. Every slice access is bounds-checked: input is untrusted.
func jpegOrientation(b []byte) int {
	if len(b) < 4 || b[0] != 0xFF || b[1] != 0xD8 {
		return 1
	}
	i := 2
	for i+4 <= len(b) {
		if b[i] != 0xFF {
			return 1
		}
		marker := b[i+1]
		switch {
		case marker == 0xFF: // fill byte
			i++
			continue
		case marker == 0xDA || marker == 0xD9: // SOS / EOI
			return 1
		case marker == 0x01 || (marker >= 0xD0 && marker <= 0xD7): // standalone
			i += 2
			continue
		}
		segLen := int(binary.BigEndian.Uint16(b[i+2 : i+4]))
		if segLen < 2 || i+2+segLen > len(b) {
			return 1
		}
		payload := b[i+4 : i+2+segLen]
		if marker == 0xE1 && len(payload) >= 6 && string(payload[:6]) == "Exif\x00\x00" {
			return tiffOrientation(payload[6:])
		}
		i += 2 + segLen
	}
	return 1
}

func tiffOrientation(t []byte) int {
	if len(t) < 8 {
		return 1
	}
	var bo binary.ByteOrder
	switch string(t[:4]) {
	case "II*\x00":
		bo = binary.LittleEndian
	case "MM\x00*":
		bo = binary.BigEndian
	default:
		return 1
	}
	off := uint64(bo.Uint32(t[4:8]))
	if off < 8 || off+2 > uint64(len(t)) {
		return 1
	}
	n := uint64(bo.Uint16(t[off : off+2]))
	for k := uint64(0); k < n; k++ {
		e := off + 2 + k*12
		if e+12 > uint64(len(t)) {
			return 1
		}
		if bo.Uint16(t[e:e+2]) != tagOrientation {
			continue
		}
		if bo.Uint16(t[e+2:e+4]) != 3 || bo.Uint32(t[e+4:e+8]) != 1 {
			return 1
		}
		v := int(bo.Uint16(t[e+8 : e+10]))
		if v < 1 || v > 8 {
			return 1
		}
		return v
	}
	return 1
}
