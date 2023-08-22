package main

import (
	"fmt"
	"log"

	"github.com/NoelM/minigo"
)

func main() {
	init := []minigo.ATCommand{
		{
			Command: "AT&F1+MCA=0",
			Reply:   "OK",
		},
		{
			Command: "AT&N2",
			Reply:   "OK",
		},
		{
			Command: "ATS27=16",
			Reply:   "OK",
		},
		{
			Command: "ATA",
			Reply:   "CONNECT 1200/75/NONE",
		},
	}

	modem := minigo.NewModem("/dev/ttyUSB0", 115200, init)

	err := modem.Init()
	if err != nil {
		log.Fatal(err)
	}

	for {
		n, buf, err := modem.Read()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(buf[:n])
	}
}
