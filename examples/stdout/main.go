package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	vdt, err := os.ReadFile("mitterand.vdt")
	os.Stdout.Write(vdt)

	file, err := os.Create("stdout.log")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	buf := []byte{}
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		file.Write([]byte(fmt.Sprintf("recv %d bytes msg='%s'\n", n, buf)))
	}
}
