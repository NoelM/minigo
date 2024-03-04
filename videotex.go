package minigo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strings"
)

func ProCode(pro byte) ([]byte, error) {
	if pro < Pro1 || pro > Pro3 {
		return nil, errors.New("pro argument beyond bound [0x39;0x3B]")
	}
	return []byte{Esc, pro}, nil
}

func PCode(i int) []byte {
	if i < 10 {
		return []byte{0x30 + byte(i)}
	} else {
		return []byte{0x30 + byte(i/10), 0x30 + byte(i%10)}
	}
}

func Word(word int) []byte {
	return []byte{GetByteHigh(word), GetByteLow(word)}
}

func CursorInScreen(x, y int, resolution uint) (bool, error) {
	switch resolution {
	case ResolutionSimple:
		return x > 0 && x <= ColonnesSimple && y > 0 && y <= LignesSimple, nil
	case ResolutionDouble:
		return x > 0 && x <= ColonnesDouble && y > 0 && y <= ColonnesSimple, nil
	default:
		return false, fmt.Errorf("unknown resolution: %d", resolution)
	}
}

func MoveAt(row, col int, csi bool) (buf []byte) {
	if row == 1 && col == 1 {
		buf = []byte{Rs}
	} else if csi && (row+col) > 12 {
		buf = Word(Csi)
		buf = append(buf, PCode(row)...)
		buf = append(buf, 0x3B)
		buf = append(buf, PCode(col)...)
		buf = append(buf, 0x48)
	} else {
		buf = []byte{Rs}
		for i := 1; i < row; i += 1 {
			buf = append(buf, Lf)
		}
		for i := 0; i < col; i += 1 {
			buf = append(buf, Ht)
		}
	}

	return
}

func MoveLeft(n int, csi bool) (buf []byte) {
	if n == 1 {
		buf = append(buf, Bs)
	} else if csi && n > 12 {
		buf = Word(Csi)
		buf = append(buf, PCode(n)...)
		buf = append(buf, 0x44)
	} else {
		for i := 0; i < n; i += 1 {
			buf = append(buf, Bs)
		}
	}
	return
}

func MoveRight(n int, csi bool) (buf []byte) {
	if n == 1 {
		buf = append(buf, Ht)
	} else if csi && n > 12 {
		buf = Word(Csi)
		buf = append(buf, PCode(n)...)
		buf = append(buf, 0x43)
	} else {
		for i := 0; i < n; i += 1 {
			buf = append(buf, Ht)
		}
	}
	return
}

func MoveDown(n int, csi bool) (buf []byte) {
	if n == 1 {
		buf = append(buf, Lf)
	} else if csi && n > 12 {
		buf = Word(Csi)
		buf = append(buf, PCode(n)...)
		buf = append(buf, 0x42)
	} else {
		for i := 0; i < n; i += 1 {
			buf = append(buf, Lf)
		}
	}
	return
}

func MoveUp(n int, csi bool) (buf []byte) {
	if n == 1 {
		buf = append(buf, Vt)
	} else if csi && n > 12 {
		buf = Word(Csi)
		buf = append(buf, PCode(n)...)
		buf = append(buf, 0x41)
	} else {
		for i := 0; i < n; i += 1 {
			buf = append(buf, Vt)
		}
	}
	return
}

func Return(n int, csi bool) (buf []byte) {
	buf = append(buf, Cr)
	buf = append(buf, MoveDown(n, csi)...)
	return
}

func ReturnUp(n int, csi bool) (buf []byte) {
	buf = append(buf, Cr)
	buf = append(buf, MoveUp(n, csi)...)
	return
}

func CleanScreen() (buf []byte) {
	buf = Word(Csi)
	buf = append(buf, 0x32, 0x4A)
	return
}

func CleanScreenFromCursor() (buf []byte) {
	buf = Word(Csi)
	buf = append(buf, 0x4A)
	return
}

func CleanScreenToCursor() (buf []byte) {
	buf = Word(Csi)
	buf = append(buf, 0x31, 0x4A)
	return
}

func CleanLine() (buf []byte) {
	buf = Word(Csi)
	buf = append(buf, 0x32, 0x4B)
	return buf
}

func CleanLineFromCursor() (buf []byte) {
	buf = Word(Csi)
	buf = append(buf, 0x4B)
	return
}

func CleanLineToCursor() (buf []byte) {
	buf = Word(Csi)
	buf = append(buf, 0x31, 0x4B)
	return
}

func CleanNItemsFromCursor(n int) (buf []byte) {
	buf = Word(Csi)
	buf = append(buf, PCode(n)...)
	buf = append(buf, 0x50)
	return
}

func CleanNRowsFromCursor(n int) (buf []byte) {
	buf = Word(Csi)
	buf = append(buf, PCode(n)...)
	buf = append(buf, 0x4D)
	return
}

func TextZone(text string, attributes ...byte) (buf []byte) {
	buf = append(buf, Sp)

	for _, atb := range attributes {
		buf = append(buf, EncodeAttribute(atb)...)
	}
	buf = append(buf, EncodeString(text)...)
	buf = append(buf, Sp)

	return
}

func SubArticle(content []byte, x, y int, res uint) (buf []byte) {
	inScreen, err := CursorInScreen(x, y, res)
	if err != nil {
		log.Printf("unable to create sub-article: %s", err.Error())
	}
	if !inScreen {
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

func RepeatRune(r rune, n int) (buf []byte) {
	if n > 40 {
		return
	}

	buf = EncodeRune(r)
	buf = append(buf, Rep)
	buf = append(buf, 0x40+byte(n))
	return
}

func HLine(row, col, len int, t LineType, csi bool) (buf []byte) {
	buf = MoveAt(row, col, csi)
	buf = append(buf, byte(t), Rep, 0x40+byte(len-1))
	return
}

func VLine(row, col, len int, t LineType, csi bool) (buf []byte) {
	buf = MoveAt(row, col, csi)

	for i := 0; i < len; i += 1 {
		// BS = moves cursor left
		// LF = moves cursor down
		buf = append(buf, byte(t), Bs, Lf)
	}
	return
}

func Rectangle(row, col, width, height int, csi bool) (buf []byte) {
	buf = HLine(row, col, width, Bottom, csi)
	buf = append(buf, VLine(row+1, col, height-2, Left, csi)...)
	buf = append(buf, VLine(row+1, col+width, height-2, Left, csi)...)
	buf = append(buf, HLine(row+height-1, col, width, Top, csi)...)
	return
}

func EncodeBytes(byt []byte) (buf []byte) {
	return EncodeString(string(byt))
}

func EncodeString(msg string) (buf []byte) {
	for _, c := range msg {
		if b := EncodeRune(c); b != nil {
			buf = append(buf, b...)
		}
	}
	return
}

func EncodeSprintf(format string, a ...any) []byte {
	return EncodeString(fmt.Sprintf(format, a...))
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

		} else if entryBytes[1] == 0x5B {
			// CSI = ESC(1B) + 5B
			if len(entryBytes) == 2 {
				return
			}

		} else if entryBytes[1] == Pro2 {
			// ESC + 0x3A (PRO2) + [2 BYTES]
			if len(entryBytes) < 4 {
				return
			}

			done, pro = true, true
			return

		} else if entryBytes[1] == Pro3 {
			// ESC + 0x3B (PRO3) + [3 BYTES]
			if len(entryBytes) < 5 {
				return
			}

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
		copy(tmp[1:], entryBytes)

		value = int32(binary.BigEndian.Uint32(tmp))
	case 4:
		value = int32(binary.BigEndian.Uint32(entryBytes))
	default:
		err = errors.New("unable to cast readbuffer")
	}

	done = true
	return
}
