package main

import (
	"fmt"
	"log"
	"os"

	"github.com/NoelM/minigo"
)

func main() {
	buf := minigo.GetCleanScreen()
	buf = append(buf, minigo.GetMoveCursorXY(1, 2)...)
	buf = append(buf, minigo.EncodeAttributes(minigo.CursorOff, minigo.DoubleGrandeur, minigo.InversionFond, minigo.Clignotement)...)
	buf = append(buf, minigo.EncodeMessage("COUCOU LE MONDE !")...)

	for id, b := range buf {
		buf[id] = minigo.GetByteWithParity(b)
	}

	os.Stdout.Write(buf)

	file, err := os.Create("stdout.log")
	if err != nil {
		log.Fatal(err)
	}

	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		file.Write([]byte(fmt.Sprintf("recv %d bytes msg='%s'\n", n, buf)))
	}
}
