package minigo

import "errors"

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
	for i := 0; i < ParityBitPosition; i++ {
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

// SetParity returns the input byte with the parity rule applied
func SetParity(b byte) byte {
	// The parity bit (at position 7) is set to:
	// * 0 -> if the sum of (0;6) bits is even,
	// * 1 -> if the sum of (0;6) bits is odd
	return BitWriteAt(b, ParityBitPosition, !IsByteEven(b))
}

// ValidAndRemoveParity, verifies the parity of the input byte:
// * if the parity is invalid, the function returns an error
// * if the parity is valid, the function returns the byte with parity bit at 0
func ValidAndRemoveParity(b byte) (byte, error) {
	// The parity bit (at position 7) is set to:
	// * 0 -> if the sum of (0;6) bits is even,
	// * 1 -> if the sum of (0;6) bits is odd
	if IsByteEven(b) != BitReadAt(b, ParityBitPosition) {
		return BitWriteAt(b, ParityBitPosition, false), nil
	} else {
		return b, errors.New("invalid parity received")
	}
}
