package minigo

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
)

const ModemReadTimeout = 30 * time.Second

type Modem struct {
	port        serial.Port
	init        []ATCommand
	ringHandler func(modem *Modem)
	connected   bool
	mutex       sync.RWMutex
}

type ATCommand struct {
	Command string
	Reply   string
}

func NewModem(portName string, baud int, init []ATCommand) (*Modem, error) {
	port, err := serial.Open(portName, &serial.Mode{BaudRate: baud})
	if err != nil {
		errorLog.Printf("unable to start modem port=%s baud=%d: %s\n", port, baud, err.Error())
		return nil, &ConnectorError{code: InvalidDefinition, raw: err}
	}

	return &Modem{
		port:      port,
		init:      init,
		connected: false,
	}, nil
}

func (m *Modem) Init() error {
	rep := strings.NewReplacer("\n", " ", "\r", " ")

	for _, at := range m.init {
		isAck, result, err := m.sendCommandAndWait(at)
		if err != nil {
			return err
		}

		if !isAck {
			return &ConnectorError{
				code: InvalidInit,
				raw:  fmt.Errorf("cannot ack command='%s' got reply='%s'", at.Command, rep.Replace(result)),
			}
		} else {
			infoLog.Printf("acknowledged command='%s' with reply='%s'", at.Command, rep.Replace(result))
		}
	}

	return nil
}

func (m *Modem) sendCommandAndWait(at ATCommand) (bool, string, error) {
	// Send initial message
	if len(at.Command) > 0 {
		if _, err := m.port.Write([]byte(at.Command + "\r\n")); err != nil {
			errorLog.Printf("unable to write to port: %s\n", err.Error())
			return false, "", &ConnectorError{code: Unreachable, raw: err}
		}
	}

	var ack bool
	var result string

	// Wait for message
	if len(at.Reply) > 0 {
		for {
			buffer, err := m.ReadTimeout(ModemReadTimeout)
			if err != nil {
				errorLog.Printf("unable to read from port: %s\n", err.Error())
				return false, "", &ConnectorError{code: Unreachable, raw: err}
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
				errorLog.Println("modem replied ERROR to command")
				break
			}
		}
	} else {
		ack = true
	}

	return ack, result, nil
}

func (m *Modem) Write(b []byte) error {
	_, err := m.port.Write(b)
	if err != nil {
		return &ConnectorError{code: Unreachable, raw: err}
	}
	return nil
}

func (m *Modem) Read() ([]byte, error) {
	buffer := make([]byte, 64)

	n, err := m.port.Read(buffer)
	if err != nil {
		return nil, &ConnectorError{code: Unreachable, raw: err}
	}
	return buffer[:n], nil
}

func (m *Modem) ReadTimeout(d time.Duration) ([]byte, error) {
	m.port.SetReadTimeout(d)
	defer m.port.SetReadTimeout(serial.NoTimeout)

	buffer := make([]byte, 64)

	n, err := m.port.Read(buffer)
	if err != nil {
		return nil, &ConnectorError{code: Unreachable, raw: err}
	}
	return buffer[:n], nil
}

func (m *Modem) Connected() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.connected
}

func (m *Modem) SetConnected(status bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.connected = status
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

		// Connection lost
		if !status.DCD && m.Connected() {
			infoLog.Println("modem lost V.23 connection")
			m.SetConnected(false)
			m.Init()
		}

		// Call recieved
		if status.RI || forceRing {
			infoLog.Println("we got a call, modem bit RING=1")
			forceRing = false
			m.Connect()

			// Fail to establish connection, reset the bouzin
			if !m.Connected() {
				m.Init()
			}
		}

		time.Sleep(time.Second)
	}

}

func (m *Modem) Connect() {
	rep := strings.NewReplacer("\n", " ", "\r", " ")

	isAck, result, err := m.sendCommandAndWait(ATCommand{Command: "ATA", Reply: "CONNECT 1200/75/NONE"})
	if err != nil {
		errorLog.Printf("unable to send and ack command: %s\n", err.Error())
		return
	}

	if !isAck {
		errorLog.Printf("unable to connect after RING got reply=%s\n", rep.Replace(result))
		return
	} else {
		infoLog.Printf("acknowledged command='ATA' with reply='%s'", rep.Replace(result))
	}

	time.Sleep(100 * time.Millisecond)

	status, err := m.port.GetModemStatusBits()
	if err != nil {
		warnLog.Printf("unable to get modem status: %s\n", err.Error())
	}

	if !status.DCD {
		errorLog.Println("unable establish connection")
		return
	}

	m.SetConnected(true)
	infoLog.Println("connection V.23 established")

	go m.ringHandler(m)
}

func (m *Modem) Disconnect() {
	if !m.Connected() {
		infoLog.Println("asked for modem disconnect, but already disconnected")
		return
	}

	infoLog.Println("switch modem connected status to false")
	m.SetConnected(false)

	m.port.SetDTR(false)
	infoLog.Println("set modem bit DTR=0, waiting 2s")
	time.Sleep(2 * time.Second)

	m.port.SetDTR(true)
	infoLog.Println("set modem bit DTR=1, waiting 2s")
	time.Sleep(2 * time.Second)

	infoLog.Println("relaunch modem init sequence")
	m.Init()
}
