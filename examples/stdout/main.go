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
	os.Stdout.Write(vdt)

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
