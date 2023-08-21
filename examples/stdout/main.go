package main

import (
	"fmt"
	"log"
	"os"

	"github.com/NoelM/minigo"
)

func main() {
	buf := minigo.GetCleanScreen()
	buf = append(buf, minigo.GetMoveCursorXY(1, 1)...)
	for id, b := range buf {
		buf[id] = minigo.GetByteWithParity(b)
	}
	os.Stdout.Write(buf)

	vdt, err := os.ReadFile("mitterrand.vdt")
	if err != nil {
		return
	}
	for _, b := range vdt {
		os.Stdout.Write([]byte{minigo.GetByteWithParity(b)})
	}

	file, err := os.Create("stdout.log")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			log.Fatal(err)
			break
		}
		file.Write([]byte(fmt.Sprintf("recv %d bytes msg='%s'\n", n, buf)))
	}
}
