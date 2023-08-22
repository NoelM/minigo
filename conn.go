package minigo

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

type Connector interface {
	Init() error

	Write([]byte) error

	Read() (int, []byte, error)

	IsClosed() bool
}

type Modem struct {
	port        serial.Port
	init        []ATCommand
	buffer      []byte
	ringHandler func(modem *Modem)
	closed      bool
}

type ATCommand struct {
	Command string
	Reply   string
}

func NewModem(portName string, baud int, init []ATCommand) *Modem {
	port, err := serial.Open(portName, &serial.Mode{BaudRate: baud})
	if err != nil {
		errorLog.Fatalf("unable to start modem port=%s baud=%d: %s", port, baud, err.Error())
	}

	return &Modem{
		port:   port,
		init:   init,
		buffer: make([]byte, 1024),
		close:  false,
	}
}

func (m *Modem) Init() error {
	for _, at := range m.init {
		if !m.sendCommandAndWait(at) {
			return fmt.Errorf("cannot ack command='%s'", at.Command)
		}
	}

	return nil
}

func (m *Modem) sendCommandAndWait(at ATCommand) bool {
	// Send initial message
	if len(at.Command) > 0 {
		if _, err := m.port.Write([]byte(at.Command + "\r\n")); err != nil {
			log.Println(err)
		}
	}

	// Wait for message
	if len(at.Reply) > 0 {
		var result string
		buffer := make([]byte, 64)
		for {
			n, err := m.port.Read(buffer)
			if err != nil {
				log.Fatalln(err)
			}
			if n == 0 {
				break
			}

			result += string(buffer[0:n])
			if strings.Contains(result, at.Reply) {
				break
			} else if strings.Contains(result, "ERROR") {
				return false
			}
		}
	}
	return true
}

func (m *Modem) Write(b []byte) error {
	_, err := m.port.Write(b)
	return err
}

func (m *Modem) Read() (int, []byte, error) {
	n, err := m.port.Read(m.buffer)
	return n, m.buffer, err
}

func (m *Modem) IsClosed() bool {
	return m.closed
}

func (m *Modem) RingHandler(f func(modem *Modem)) {
	m.ringHandler = f
}

func (m *Modem) Serve() {
	var err error
	var status *serial.ModemStatusBits

	for {
		status, err = m.port.GetModemStatusBits()
		if err != nil {
			warnLog.Printf("unable to get modem status: %s\n", err.Error())
		}
		if status.RI {
			break
		}
		time.Sleep(time.Second)
	}

	m.Connect()
}

func (m *Modem) Connect() {
	if !m.sendCommandAndWait(ATCommand{Command: "ATA", Reply: "CONNECT 1200/75/NONE"}) {
		errorLog.Fatalf("unable to connect after Ring")
	}

	go m.ringHandler(m)

	var err error
	var status *serial.ModemStatusBits
	for {
		status, err = m.port.GetModemStatusBits()
		if err != nil {
			warnLog.Printf("unable to get modem status: %s\n", err.Error())
		}
		if !status.DCD {
			break
		}
		time.Sleep(time.Second)
	}

	m.closed = true
}
