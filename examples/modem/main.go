package main

import (
	"log"
	"strings"

	"go.bug.st/serial"
)

const Port = "/dev/ttyUSB0"
const Message = "AT&F1+MCA=0"
const Wait = "OK"

func main() {
	connection, err := serial.Open(Port, &serial.Mode{BaudRate: 115200})
	if err != nil {
		log.Fatalln(err)
	}

	// Send initial message
	if len(Message) > 0 {
		if _, err := connection.Write([]byte(Message + "\r\n")); err != nil {
			log.Println(err)
		}
	}

	// Wait for message
	if len(Wait) > 0 {
		var result string
		buffer := make([]byte, 64)
		for {
			n, err := connection.Read(buffer)
			if err != nil {
				log.Fatalln(err)
			}
			if n == 0 {
				break
			}
			result += string(buffer[0:n])
			if strings.Contains(result, Wait) {
				break
			}
		}
	}
}
