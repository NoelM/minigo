package main

import (
	"os"

	"github.com/NoelM/minigo"
)

func main() {
	buf := minigo.GetCleanScreen()
	buf = append(buf, minigo.EncodeMessage("COUCOU LE MONDE !")...)

	for id, b := range buf {
		buf[id] = minigo.GetByteWithParity(b)
	}

	os.Stdout.Write(buf)
}
