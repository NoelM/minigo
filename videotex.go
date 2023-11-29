package minigo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strings"
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

func GetMoveCursorAt(row, col int) (buf []byte) {
	buf = GetWord(Csi)
	buf = append(buf, GetPCode(row)...)
	buf = append(buf, 0x3B)
	buf = append(buf, GetPCode(col)...)
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

func GetCleanNItemsFromCursor(n int) (buf []byte) {
	buf = GetWord(Csi)
	buf = append(buf, GetPCode(n)...)
	buf = append(buf, 0x50)
	return
}

func GetCleanNRowsFromCursor(n int) (buf []byte) {
	buf = GetWord(Csi)
	buf = append(buf, GetPCode(n)...)
	buf = append(buf, 0x4D)
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

func EncodeCharToVideotex(c byte) byte {
	return byte(strings.LastIndexByte(CharTable, c))
}

func EncodeSpecial(r rune) []byte {
	switch r {
	case '’':
		return []byte{'\''}
	case 'à':
		return []byte{Ss2, AccentGrave, 'a'}
	case 'À':
		return []byte{Ss2, AccentGrave, 'A'}
	case 'â':
		return []byte{Ss2, AccentCirconflexe, 'a'}
	case 'ä':
		return []byte{Ss2, Trema, 'a'}
	case 'è':
		return []byte{Ss2, AccentGrave, 'e'}
	case 'È':
		return []byte{Ss2, AccentGrave, 'E'}
	case 'é':
		return []byte{Ss2, AccentAigu, 'e'}
	case 'É':
		return []byte{Ss2, AccentAigu, 'E'}
	case 'ê':
		return []byte{Ss2, AccentCirconflexe, 'e'}
	case 'ë':
		return []byte{Ss2, Trema, 'e'}
	case 'î':
		return []byte{Ss2, AccentCirconflexe, 'i'}
	case 'ï':
		return []byte{Ss2, Trema, 'i'}
	case 'ö':
		return []byte{Ss2, Trema, 'o'}
	case 'ô':
		return []byte{Ss2, AccentCirconflexe, 'o'}
	case 'ù':
		return []byte{Ss2, AccentGrave, 'u'}
	case 'û':
		return []byte{Ss2, AccentCirconflexe, 'u'}
	case 'ü':
		return []byte{Ss2, Trema, 'u'}
	case 'ç':
		return []byte{Ss2, Cedille, 'c'}
	case 'Ç':
		return []byte{Ss2, Cedille, 'C'}
	case '£':
		return []byte{Ss2, Livre}
	case '$':
		return []byte{Ss2, Dollar}
	case '#':
		return []byte{Ss2, Diese}
	case '§':
		return []byte{Ss2, Paragraphe}
	case '←':
		return []byte{Ss2, FlecheGauche}
	case '↑':
		return []byte{Ss2, FlecheHaut}
	case '→':
		return []byte{Ss2, FlecheDroite}
	case '↓':
		return []byte{Ss2, FlecheBas}
	case '°':
		return []byte{Ss2, Degre}
	case '±':
		return []byte{Ss2, PlusOuMoins}
	case '÷':
		return []byte{Ss2, Division}
	case '¼':
		return []byte{Ss2, UnQuart}
	case '½':
		return []byte{Ss2, UnDemi}
	case '¾':
		return []byte{Ss2, TroisQuart}
	case 'œ':
		return []byte{Ss2, OeMinuscule}
	case 'Œ':
		return []byte{Ss2, OeMajuscule}
	case 'ß':
		return []byte{Ss2, Beta}
	}

	return nil
}

func IsAccent(b byte) bool {
	return b == AccentAigu ||
		b == AccentGrave ||
		b == AccentCirconflexe ||
		b == Trema ||
		b == Cedille
}

func DecodeAccent(keyBuffer []byte) rune {
	accent := keyBuffer[0]
	letter := keyBuffer[1]

	if accent == AccentAigu {
		if letter == 'e' {
			return 'é'
		} else if letter == 'E' {
			return 'É'
		}

	} else if accent == AccentGrave {
		if letter == 'a' {
			return 'à'
		} else if letter == 'A' {
			return 'À'
		} else if letter == 'e' {
			return 'è'
		} else if letter == 'E' {
			return 'È'
		} else if letter == 'u' {
			return 'ù'
		} else if letter == 'U' {
			return 'Ù'
		}

	} else if accent == AccentCirconflexe {
		if letter == 'a' {
			return 'â'
		} else if letter == 'e' {
			return 'ê'
		} else if letter == 'i' {
			return 'î'
		} else if letter == 'o' {
			return 'ô'
		} else if letter == 'u' {
			return 'û'
		}

	} else if accent == Trema {
		if letter == 'a' {
			return 'ä'
		} else if letter == 'e' {
			return 'ë'
		} else if letter == 'i' {
			return 'ï'
		} else if letter == 'o' {
			return 'ö'
		} else if letter == 'u' {
			return 'ü'
		}

	} else if accent == Cedille {
		if letter == 'c' {
			return 'ç'
		} else if letter == 'C' {
			return 'Ç'
		}

	}

	return 0
}

func ValidChar(c byte) bool {
	return c >= Sp && c <= Del
}

func EncodeRune(r rune) []byte {
	if specialRune := EncodeSpecial(r); specialRune != nil {
		return specialRune
	}

	vdtByte := EncodeCharToVideotex(byte(r))
	if ValidChar(vdtByte) {
		return []byte{vdtByte}
	}

	return nil
}

func ValidRune(r rune) bool {
	return EncodeRune(r) != nil
}

func GetRepeatRune(r rune, n int) (buf []byte) {
	if n > 40 {
		return
	}

	buf = EncodeRune(r)
	buf = append(buf, Rep)
	buf = append(buf, 0x40+byte(n))
	return
}

func EncodeMessage(msg string) (buf []byte) {
	for _, c := range msg {
		if b := EncodeRune(c); b != nil {
			buf = append(buf, b...)
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

func ApplyParity(in []byte) (out []byte) {
	out = make([]byte, len(in))

	for id, b := range in {
		out[id] = GetByteWithParity(b)
	}

	return out
}

func ReadEntryBytes(entryBytes []byte) (done bool, pro bool, value int32, err error) {
	if entryBytes[0] == Ss2 {
		// Special characters, switch G2 mode
		if len(entryBytes) <= 1 {
			return
		}

		switch entryBytes[1] {
		case Livre:
			value = '£'
		case Paragraphe:
			value = '§'
		case Degre:
			value = '°'
		case PlusOuMoins:
			value = '±'
		case Division:
			value = '÷'
		case Beta:
			value = 'ß'
		default:
			if IsAccent(entryBytes[1]) {
				if len(entryBytes) <= 2 {
					return
				}

				// Ignore the SS2 header
				value = DecodeAccent(entryBytes[1:])
			}
		}

		if value != 0 {
			done = true
			return
		}

	} else if entryBytes[0] == Special {
		if len(entryBytes) == 1 {
			return
		}

	} else if entryBytes[0] == Prog {
		if len(entryBytes) == 1 {
			return
		}

		if entryBytes[1] == RepStatusClavier {
			if len(entryBytes) == 2 {
				return
			}

			if entryBytes[2] == CodeReceptionClavier {
				if len(entryBytes) == 3 {
					return
				}

				done, pro = true, true
				return
			}
		}

	} else if entryBytes[0] == Esc {
		if len(entryBytes) == 1 {
			return
		}

		if entryBytes[1] == CodeReceptionPrise {
			if len(entryBytes) == 2 {
				return
			}

			if entryBytes[2] == 0x34 || entryBytes[2] == 0x32 {
				if len(entryBytes) == 3 {
					return
				}
			}

		} else if entryBytes[1] == 0x5B { // CSI = ESC(1B) + 5B
			if len(entryBytes) == 2 {
				return
			}

		} else if entryBytes[1] == Pro2 { // PRO2 = ESC + 0x3A
			if len(entryBytes) < 4 {
				return
			}
			// PRO2, RESP BYTE, STATUS BYTE
			done, pro = true, true
			return
		}
	}

	switch len(entryBytes) {
	case 1:
		value = int32(entryBytes[0])
	case 2:
		value = int32(binary.BigEndian.Uint16(entryBytes))
	case 3:
		tmp := make([]byte, 4)
		copy(tmp[1:3], entryBytes)

		value = int32(binary.BigEndian.Uint32(tmp))
	case 4:
		value = int32(binary.BigEndian.Uint32(entryBytes))
	default:
		err = errors.New("unable to cast readbuffer")
	}

	done = true
	return
}
