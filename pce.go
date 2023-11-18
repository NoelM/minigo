package minigo

import "github.com/snksoft/crc"

const VDTPoly = 0b10001001

var VDTCRC = crc.Parameters{
	Width:      8,
	Polynomial: VDTPoly,
	Init:       0x0,
	ReflectIn:  false,
	ReflectOut: false,
	FinalXor:   0x0,
}

func ComputePCEBlock(buf []byte) []byte {
	// make sure we a have the good length
	inner := make([]byte, 15)
	copy(inner, buf)

	result := crc.CalculateCRC(&VDTCRC, inner)
	inner = append(inner, GetByteWithParity(byte(result)), 0x0)

	return inner
}
