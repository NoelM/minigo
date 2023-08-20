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

func GetByteWithParity(b byte) byte {
	// The parity bit is set to 0 if the sum of other bits is even,
	// thus if the sum is odd the parity bit is set to 1
	return BitWriteAt(b, ParityBitPosition, !IsByteEven(b))
}

func CheckByteParity(b byte) (byte, error) {
	// The parity bit is set to 0 if the sum of other bits is even,
	// thus if the sum is odd the parity bit is set to 1
	if IsByteEven(b) != BitReadAt(b, ParityBitPosition) {
		return BitWriteAt(b, ParityBitPosition, false), nil
	} else {
		return 0, errors.New("invalid parity received")
	}
}
