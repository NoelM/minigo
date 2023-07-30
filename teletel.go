package minigo

import (
	"errors"
	"fmt"
	"log"
)

const ByteParityPos = 7

func BitReadAt(b byte, i int) bool {
	return b&byte(1<<i) > 0
}

func GetByteLow(w int) byte {
	return byte(w & 0xFF)
}

func GetByteHigh(w int) byte {
	return byte(w >> 8)
}

func IsByteEven(b byte) bool {
	even := true
	for i := 0; i < ByteParityPos; i++ {
		if BitReadAt(b, i) {
			even = !even
		}
	}
	return even
}

func BitWriteAt(b byte, pos int, value bool) byte {
	if value {
		return b | byte(1<<pos)
	} else {
		return b &^ byte(1<<pos)
	}
}

func GetByteWithParity(b byte) byte {
	// The parity bit is set to 0 if the sum of other bits is even,
	// thus if the sum is odd the parity bit is set to 1
	return BitWriteAt(b, ByteParityPos, !IsByteEven(b))
}

func CheckByteParity(b byte) (byte, error) {
	// The parity bit is set to 0 if the sum of other bits is even,
	// thus if the sum is odd the parity bit is set to 1
	if IsByteEven(b) != BitReadAt(b, ByteParityPos) {
		return BitWriteAt(b, ByteParityPos, false), nil
	} else {
		return 0, errors.New("invalid parity received")
	}
}

func GetProCode(pro byte) ([]byte, error) {
	if pro < Pro1 || pro > Pro3 {
		return nil, errors.New("pro argument beyond bound [0x39;0x3B]")
	}
	return []byte{GetByteWithParity(Esc), GetByteWithParity(pro)}, nil
}

func GetPCode(i int) []byte {
	if i < 10 {
		return []byte{GetByteWithParity(0x30 + byte(i))}
	} else {
		return []byte{GetByteWithParity(0x30 + byte(i/10)), GetByteWithParity(0x30 + byte(i%10))}
	}
}

func GetWordWithParity(word int) []byte {
	return []byte{GetByteWithParity(GetByteHigh(word)), GetByteWithParity(GetByteLow(word))}
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
	buf = GetWordWithParity(Csi)
	buf = append(buf, GetPCode(y)...)
	buf = append(buf, GetByteWithParity(0x3B))
	buf = append(buf, GetPCode(x)...)
	buf = append(buf, GetByteWithParity(0x48))
	return
}

func GetMoveCursorLeft(n int) (buf []byte) {
	if n == 1 {
		buf = append(buf, GetByteWithParity(Bs))
	} else {
		buf = GetWordWithParity(Csi)
		buf = append(buf, GetPCode(n)...)
		buf = append(buf, GetByteWithParity(0x44))
	}
	return
}

func GetMoveCursorRight(n int) (buf []byte) {
	if n == 1 {
		buf = append(buf, GetByteWithParity(Ht))
	} else {
		buf = GetWordWithParity(Csi)
		buf = append(buf, GetPCode(n)...)
		buf = append(buf, GetByteWithParity(0x43))
	}
	return
}

func GetMoveCursorDown(n int) (buf []byte) {
	if n == 1 {
		buf = append(buf, GetByteWithParity(Lf))
	} else {
		buf = GetWordWithParity(Csi)
		buf = append(buf, GetPCode(n)...)
		buf = append(buf, GetByteWithParity(0x42))
	}
	return
}

func GetMoveCursorUp(n int) (buf []byte) {
	if n == 1 {
		buf = append(buf, GetByteWithParity(Vt))
	} else {
		buf = GetWordWithParity(Csi)
		buf = append(buf, GetPCode(n)...)
		buf = append(buf, GetByteWithParity(0x41))
	}
	return
}

func GetMoveCursorReturn(n int) (buf []byte) {
	buf = append(buf, GetByteWithParity(Cr))
	buf = append(buf, GetMoveCursorDown(n)...)
	return
}

func GetCleanScreen() (buf []byte) {
	buf = GetWordWithParity(Csi)
	buf = append(buf, GetByteWithParity(0x32), GetByteWithParity(0x4A))
	return
}

func GetCleanScreenFromCursor() (buf []byte) {
	buf = GetWordWithParity(Csi)
	buf = append(buf, GetByteWithParity(0x4A))
	return
}

func GetCleanScreenToCursor() (buf []byte) {
	buf = GetWordWithParity(Csi)
	buf = append(buf, GetByteWithParity(0x31), GetByteWithParity(0x4A))
	return
}

func GetCleanLine() (buf []byte) {
	buf = GetWordWithParity(Csi)
	buf = append(buf, GetByteWithParity(0x32), GetByteWithParity(0x4B))
	return buf
}

func GetCleanLineFromCursor() (buf []byte) {
	buf = GetWordWithParity(Csi)
	buf = append(buf, GetByteWithParity(0x4B))
	return
}

func GetCleanLineToCursor() (buf []byte) {
	buf = GetWordWithParity(Csi)
	buf = append(buf, GetByteWithParity(0x31), GetByteWithParity(0x4B))
	return
}

func EncodeChar(c int32) (byte, error) {
	vdtByte := GetVideotextCharByte(byte(c))
	if IsValidChar(vdtByte) {
		return vdtByte, nil
	}
	return 0, errors.New("invalid char byte")
}

func EncodeMessage(msg string) (buf []byte) {
	for _, c := range msg {
		if b, err := EncodeChar(c); err == nil {
			buf = append(buf, GetByteWithParity(b))
		} else {
			continue
		}
	}
	return
}

func EncodeAttribute(attribute byte) (buf []byte) {
	buf = append(buf, GetByteWithParity(Esc), GetByteWithParity(attribute))
	return
}

func EncodeAttributes(attributes ...byte) (buf []byte) {
	for _, atb := range attributes {
		buf = append(buf, GetByteWithParity(Esc), GetByteWithParity(atb))
	}
	return
}

func GetTextZone(text string, attributes ...byte) (buf []byte) {
	buf = append(buf, GetByteWithParity(Sp))

	for _, atb := range attributes {
		buf = append(buf, EncodeAttribute(atb)...)
	}
	buf = append(buf, EncodeMessage(text)...)
	buf = append(buf, GetByteWithParity(Sp))

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

	buf = append(buf, GetByteWithParity(Us), byte(0x40+x), byte(0x40+y))
	buf = append(buf, content...)
	return
}

func GetCursorOn() byte {
	return GetByteWithParity(CursorOn)
}

func AppendCursorOff() byte {
	return GetByteWithParity(CursorOff)
}
