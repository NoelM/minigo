package minigo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

func GetProCode(pro byte) ([]byte, error) {
	if pro < Pro1 || pro > Pro3 {
		return nil, errors.New("pro argument beyond bound [0x39;0x3B]")
	}
	return []byte{Esc, pro}, nil
}

func GetPCode(i int) []byte {
	if i < 10 {
		return []byte{0x30 + byte(i)}
	} else {
		return []byte{0x30 + byte(i/10), 0x30 + byte(i%10)}
	}
}

func GetWord(word int) []byte {
	return []byte{GetByteHigh(word), GetByteLow(word)}
}

func IsPosInBounds(x, y int, resolution uint) (bool, error) {
	switch resolution {
	case ResolutionSimple:
		return x > 0 && x <= ColonnesSimple && y > 0 && y <= LignesSimple, nil
	case ResolutionDouble:
		return x > 0 && x <= ColonnesDouble && y > 0 && y <= ColonnesSimple, nil
	default:
		return false, fmt.Errorf("unknown resolution: %d", resolution)
	}
}

func GetMoveCursorXY(x, y int) (buf []byte) {
	buf = GetWord(Csi)
	buf = append(buf, GetPCode(y)...)
	buf = append(buf, 0x3B)
	buf = append(buf, GetPCode(x)...)
	buf = append(buf, 0x48)
	return
}

func GetMoveCursorLeft(n int) (buf []byte) {
	if n == 1 {
		buf = append(buf, Bs)
	} else {
		buf = GetWord(Csi)
		buf = append(buf, GetPCode(n)...)
		buf = append(buf, 0x44)
	}
	return
}

func GetMoveCursorRight(n int) (buf []byte) {
	if n == 1 {
		buf = append(buf, Ht)
	} else {
		buf = GetWord(Csi)
		buf = append(buf, GetPCode(n)...)
		buf = append(buf, 0x43)
	}
	return
}

func GetMoveCursorDown(n int) (buf []byte) {
	if n == 1 {
		buf = append(buf, Lf)
	} else {
		buf = GetWord(Csi)
		buf = append(buf, GetPCode(n)...)
		buf = append(buf, 0x42)
	}
	return
}

func GetMoveCursorUp(n int) (buf []byte) {
	if n == 1 {
		buf = append(buf, Vt)
	} else {
		buf = GetWord(Csi)
		buf = append(buf, GetPCode(n)...)
		buf = append(buf, 0x41)
	}
	return
}

func GetMoveCursorReturn(n int) (buf []byte) {
	buf = append(buf, Cr)
	buf = append(buf, GetMoveCursorDown(n)...)
	return
}

func GetCleanScreen() (buf []byte) {
	buf = GetWord(Csi)
	buf = append(buf, 0x32, 0x4A)
	return
}

func GetCleanScreenFromCursor() (buf []byte) {
	buf = GetWord(Csi)
	buf = append(buf, 0x4A)
	return
}

func GetCleanScreenToCursor() (buf []byte) {
	buf = GetWord(Csi)
	buf = append(buf, 0x31, 0x4A)
	return
}

func GetCleanLine() (buf []byte) {
	buf = GetWord(Csi)
	buf = append(buf, 0x32, 0x4B)
	return buf
}

func GetCleanLineFromCursor() (buf []byte) {
	buf = GetWord(Csi)
	buf = append(buf, 0x4B)
	return
}

func GetCleanLineToCursor() (buf []byte) {
	buf = GetWord(Csi)
	buf = append(buf, 0x31, 0x4B)
	return
}

func EncodeChar(c int32) (byte, error) {
	vdtByte := GetVideotextCharByte(byte(c))
	if IsByteAValidChar(vdtByte) {
		return vdtByte, nil
	}
	return 0, errors.New("invalid char byte")
}

func EncodeMessage(msg string) (buf []byte) {
	for _, c := range msg {
		if b, err := EncodeChar(c); err == nil {
			buf = append(buf, b)
		} else {
			continue
		}
	}
	return
}

func EncodeSprintf(format string, a ...any) []byte {
	return EncodeMessage(fmt.Sprintf(format, a...))
}

func EncodeAttribute(attribute byte) (buf []byte) {
	buf = append(buf, Esc, attribute)
	return
}

func EncodeAttributes(attributes ...byte) (buf []byte) {
	for _, atb := range attributes {
		buf = append(buf, EncodeAttribute(atb)...)
	}
	return
}

func GetTextZone(text string, attributes ...byte) (buf []byte) {
	buf = append(buf, Sp)

	for _, atb := range attributes {
		buf = append(buf, EncodeAttribute(atb)...)
	}
	buf = append(buf, EncodeMessage(text)...)
	buf = append(buf, Sp)

	return
}

func GetSubArticle(content []byte, x, y int, res uint) (buf []byte) {
	inBound, err := IsPosInBounds(x, y, res)
	if err != nil {
		log.Printf("unable to create sub-article: %s", err.Error())
	}
	if !inBound {
		log.Printf("positon (x=%d ; y=%d) out-of-bounds", x, y)
	}

	buf = append(buf, Us, byte(0x40+x), byte(0x40+y))
	buf = append(buf, content...)
	return
}

func GetCursorOn() byte {
	return CursorOn
}

func GetCursorOff() byte {
	return CursorOff
}

func ReadKey(keyBuffer []byte) (done bool, pro bool, value uint, err error) {
	if keyBuffer[0] == 0x19 {
		if len(keyBuffer) == 1 {
			return
		}

		switch keyBuffer[1] {
		case 0x23:
			keyBuffer = []byte{0xA3}
		case 0x27:
			keyBuffer = []byte{0xA7}
		case 0x30:
			keyBuffer = []byte{0xB0}
		case 0x31:
			keyBuffer = []byte{0xB1}
		case 0x38:
			keyBuffer = []byte{0xF7}
		case 0x7B:
			keyBuffer = []byte{0xDF}
		}
	} else if keyBuffer[0] == 0x13 {
		if len(keyBuffer) == 1 {
			return
		}
	} else if keyBuffer[0] == Esc {
		if len(keyBuffer) == 1 {
			return
		}

		if keyBuffer[1] == 0x5B {
			if len(keyBuffer) == 2 {
				return
			}

			if keyBuffer[2] == 0x34 || keyBuffer[2] == 0x32 {
				if len(keyBuffer) == 3 {
					return
				}
			}
		} else if keyBuffer[1] == Pro2 { // PRO2 = ESC + 0x3A
			if len(keyBuffer) < 4 {
				return
			}
			// PRO2, RESP BYTE, STATUS BYTE
			done, pro = true, true
			return
		}
	}

	switch len(keyBuffer) {
	case 1:
		value = uint(keyBuffer[0])
	case 2:
		value = uint(binary.BigEndian.Uint16(keyBuffer))
	case 4:
		value = uint(binary.BigEndian.Uint32(keyBuffer))
	case 8:
		value = uint(binary.BigEndian.Uint64(keyBuffer))
	default:
		err = errors.New("unable to cast readbuffer")
	}

	done = true
	return
}
