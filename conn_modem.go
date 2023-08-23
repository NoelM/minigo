package minigo

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

const ModemReadTimeout = 60 * time.Second

type Modem struct {
	port        serial.Port
	init        []ATCommand
	ringHandler func(modem *Modem)
	connected   bool
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
		port:      port,
		init:      init,
		connected: false,
	}
}

func (m *Modem) Init() error {
	rep := strings.NewReplacer("\n", " ", "\r", " ")

	for _, at := range m.init {
		isAck, result := m.sendCommandAndWait(at)
		if !isAck {
			return fmt.Errorf("cannot ack command='%s' got reply='%s'", at.Command, rep.Replace(result))
		} else {
			infoLog.Printf("acknowledged command='%s' with reply='%s'", at.Command, rep.Replace(result))
		}
	}

	return nil
}

func (m *Modem) sendCommandAndWait(at ATCommand) (ack bool, result string) {
	// Send initial message
	if len(at.Command) > 0 {
		if _, err := m.port.Write([]byte(at.Command + "\r\n")); err != nil {
			log.Println(err)
		}
	}

	ack = false
	// Wait for message
	if len(at.Reply) > 0 {
		for {
			buffer, err := m.ReadTimeout(ModemReadTimeout)
			if err != nil {
				errorLog.Printf("read timeout error: %s\n", err.Error())
				break
			}
			if len(buffer) == 0 {
				warnLog.Println("no data replied")
				break
			}

			result += string(buffer)
			if strings.Contains(result, at.Reply) {
				ack = true
				break
			} else if strings.Contains(result, "ERROR") {
				errorLog.Println("ERROR replied by the Modem")
				break
			}
		}
	} else {
		ack = true
	}

	return
}

func (m *Modem) Write(b []byte) error {
	_, err := m.port.Write(b)
	return err
}

func (m *Modem) Read() ([]byte, error) {
	buffer := make([]byte, 64)

	n, err := m.port.Read(buffer)
	return buffer[:n], err
}

func (m *Modem) ReadTimeout(d time.Duration) ([]byte, error) {
	m.port.SetReadTimeout(d)
	defer m.port.SetReadTimeout(serial.NoTimeout)

	buffer := make([]byte, 64)

	n, err := m.port.Read(buffer)
	return buffer[:n], err
}

func (m *Modem) Connected() bool {
	return m.connected
}

func (m *Modem) RingHandler(f func(modem *Modem)) {
	m.ringHandler = f
}

func (m *Modem) Serve(forceRing bool) {
	var err error
	var status *serial.ModemStatusBits

	for {
		status, err = m.port.GetModemStatusBits()
		if err != nil {
			warnLog.Printf("unable to get modem status: %s\n", err.Error())
		}

		if !status.DCD && m.connected {
			infoLog.Println("closed connection")
			m.connected = false
			m.Init()
		}

		if status.RI || forceRing {
			infoLog.Println("RING=1, phone rings")
			forceRing = false
			m.Connect()
		}

		time.Sleep(time.Second)
	}

}

func (m *Modem) Connect() {
	rep := strings.NewReplacer("\n", " ", "\r", " ")

	isAck, result := m.sendCommandAndWait(ATCommand{Command: "ATA", Reply: "CONNECT 1200/75/NONE"})
	if !isAck {
		errorLog.Printf("unable to connect after RING got reply=%s\n", rep.Replace(result))
		return
	}

	status, err := m.port.GetModemStatusBits()
	if err != nil {
		warnLog.Printf("unable to get modem status: %s\n", err.Error())
	}
	if status.DCD {
		m.connected = true
		infoLog.Println("connection V.23 established")
	} else {
		errorLog.Println("unable establish connection")
		return
	}

	go m.ringHandler(m)
}
