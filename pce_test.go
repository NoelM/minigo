package minigo

import "testing"

func TestCRC(t *testing.T) {
	data := []byte{0x03}
	expCrc := byte(0x1b)

	crc := getCRC7(data)

	if expCrc != crc {
		t.Fatalf("incoherent crc: exp=%b cur=%b\n", expCrc, crc)
	}
}
