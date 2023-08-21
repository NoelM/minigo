package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/NoelM/minigo"
)

func main() {
	file, err := os.Create("/var/log/notel/stdout.log")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()

	buf := minigo.GetCleanScreen()
	buf = append(buf, minigo.GetMoveCursorXY(1, 1)...)
	buf = append(buf, minigo.EncodeMessage("BIENVENUE, ECRIVEZ VOTRE MESSAGE")...)
	buf = append(buf, minigo.GetMoveCursorReturn(1)...)
	file.Write([]byte("buf OK\n"))

	for id, b := range buf {
		buf[id] = minigo.GetByteWithParity(b)
	}
	os.Stdout.Write(buf)
	file.Write([]byte("parity OK\n"))

	buf = []byte{}
	file.Write([]byte("start listen\n"))
	reader := bufio.NewReader(os.Stdin)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			file.Write([]byte(fmt.Sprintf("error while reading: %s\n", err.Error())))
			break
		}
		file.Write([]byte(fmt.Sprintf("recv %d bytes msg='%s'\n", n, buf[:n])))
	}
	file.Write([]byte("byte bye\n"))
}
