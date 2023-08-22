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
	}

	modem := minigo.NewModem("/dev/ttyUSB0", 115200, init)

	err := modem.Init()
	if err != nil {
		log.Fatal(err)
	}

	modem.RingHandler(func(m *minigo.Modem) {
		for m.Connected() {
			n, buf, err := m.Read()
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(buf[:n])
		}
	})

	modem.Serve(true)
}
