package main

import (
	"fmt"
	"log"
	"strings"

	"go.bug.st/serial"
)

func main() {
	port, err := serial.Open("/dev/ttyUSB0", &serial.Mode{BaudRate: 115200})
	if err != nil {
		log.Fatalln(err)
	}

	init := make([][]string, 0)
	init = append(init, []string{"AT&F1+MCA=0", "OK"})
	init = append(init, []string{"AT&N2", "OK"})
	init = append(init, []string{"ATS27=16", "OK"})
	init = append(init, []string{"ATA", ""})

	for _, pair := range init {
		SendCommandAndWait(port, pair[0], pair[1])
	}
}

func SendCommandAndWait(port serial.Port, command, reply string) {
	// Send initial message
	if len(command) > 0 {
		if _, err := port.Write([]byte(command + "\r\n")); err != nil {
			log.Println(err)
		}
	}

	// Wait for message
	if len(reply) > 0 {
		var result string
		buffer := make([]byte, 64)
		for {
			n, err := port.Read(buffer)
			if err != nil {
				log.Fatalln(err)
			}
			if n == 0 {
				break
			}
			result += string(buffer[0:n])
			if strings.Contains(result, reply) {
				break
			}
		}
		fmt.Println(result)
	}
}
