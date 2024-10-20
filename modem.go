package minigo

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"go.bug.st/serial"
)

const ModemInitTimeout = 5 * time.Second
const ModemConnectTimeout = 60 * time.Second
const ModemServeTimeout = 10 * time.Millisecond

type Modem struct {
	port        serial.Port
	init        []ATCommand
	ringHandler func(modem *Modem)
	connected   bool
	mutex       sync.RWMutex

	tag         string
	connAttempt *prometheus.CounterVec
}

type ATCommand struct {
	Command string
	Reply   string
}

func NewModem(portName string, baud int, init []ATCommand, tag string, connAttempt *prometheus.CounterVec) (*Modem, error) {
	port, err := serial.Open(portName, &serial.Mode{BaudRate: baud})
	if err != nil {
		errorLog.Printf("unable to start modem port=%s baud=%d: %s\n", port, baud, err.Error())
		return nil, &ConnectorError{code: InvalidDefinition, raw: err}
	}

	return &Modem{
		port:        port,
		init:        init,
		connected:   false,
		tag:         tag,
		connAttempt: connAttempt,
	}, nil
}

func (m *Modem) Init() error {
	infoLog.Printf("[modem] %s: init sequence", m.tag)
	m.resetBuffers()
	m.switchDTR()

	// We expect a fast reply from the modem
	m.port.SetReadTimeout(ModemInitTimeout)

	rep := strings.NewReplacer("\n", " ", "\r", " ")
	for _, atCommand := range m.init {
		infoLog.Printf("[modem] %s: send command='%s'\n", m.tag, atCommand)

		isAck, result, err := m.sendCommandAndAck(atCommand)
		if err != nil {
			return err
		}

		if !isAck {
			return &ConnectorError{
				code: InvalidInit,
				raw:  fmt.Errorf("[modem] %s: cannot ack command='%s' got reply='%s'", m.tag, atCommand.Command, rep.Replace(result)),
			}
		} else {
			infoLog.Printf("[modem] %s: acknowledged command='%s' with reply='%s'", m.tag, atCommand.Command, rep.Replace(result))
		}
	}

	return nil
}

func (m *Modem) sendCommandAndAck(atCmd ATCommand) (bool, string, error) {
	// Send initial message
	if len(atCmd.Command) > 0 {
		if _, err := m.port.Write([]byte(atCmd.Command + "\r\n")); err != nil {
			errorLog.Printf("[modem] %s: unable to write to port=%s\n", m.tag, err.Error())

			return false, "", &ConnectorError{code: Unreachable, raw: err}
		}
	}

	var ack bool
	var result string

	// Wait for message
	if len(atCmd.Reply) > 0 {
		for {
			buffer, err := m.Read()
			if err != nil {
				errorLog.Printf("[modem] %s: unable to read from port=%s\n", m.tag, err.Error())
				return false, "", &ConnectorError{code: Unreachable, raw: err}
			}
			if len(buffer) == 0 {
				warnLog.Printf("[modem] %s: no data replied\n", m.tag)
				break
			}

			result += string(buffer)
			if strings.Contains(result, atCmd.Reply) {
				ack = true
				break
			} else if strings.Contains(result, "ERROR") {
				errorLog.Printf("[modem] %s: modem replied ERROR\n", m.tag)
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
	buffer := make([]byte, 256)

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
			warnLog.Printf("[modem] %s: unable to get modem status: %s\n", m.tag, err.Error())
		}

		// Connection lost
		if !status.DCD && m.Connected() {
			warnLog.Printf("[modem] %s: lost connection\n", m.tag)

			m.switchDTR()
			m.SetConnected(false)
			m.Init()
		}

		// Call recieved
		if status.RI || forceRing {
			infoLog.Printf("[modem] %s: we got a call\n", m.tag)
			forceRing = false

			m.connAttempt.With(prometheus.Labels{"source": m.tag}).Inc()
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
	infoLog.Printf("[modem] %s: start up connection procedure\n", m.tag)
	m.port.SetReadTimeout(ModemServeTimeout)

	var status *serial.ModemStatusBits
	var result string

	start := time.Now()
	for time.Since(start) < ModemConnectTimeout {
		status, _ = m.port.GetModemStatusBits()
		time.Sleep(1 * time.Second)

		if status.DCD {
			m.SetConnected(true)
			infoLog.Printf("[modem] %s: connection established\n", m.tag)

			break
		}
	}

	if !m.Connected() {
		m.switchDTR()

		rep := strings.NewReplacer("\n", " ", "\r", " ")
		errorLog.Printf("[modem] %s: unable to connect after RING got reply=%s\n", m.tag, rep.Replace(result))
		return
	}

	m.resetBuffers()

	// Start to serve the teletel content
	go m.ringHandler(m)
}

func (m *Modem) Disconnect() {
	if !m.Connected() {
		infoLog.Printf("[modem] %s: request modem disconnect, but already disconnected\n", m.tag)
		return
	}

	m.switchDTR()
	m.SetConnected(false)
	infoLog.Printf("[modem] %s: modem disconnected\n", m.tag)
}

func (m *Modem) switchDTR() {
	m.port.SetDTR(false)
	time.Sleep(2 * time.Second)
	m.port.SetDTR(true)
}

func (m *Modem) resetBuffers() {
	// Cleanup all the In/Out buffers between the DCE and the DTE
	if err := m.port.ResetInputBuffer(); err != nil {
		errorLog.Printf("[modem] %s:unable to reset input buffer: %s\n", m.tag, err.Error())
	}
	if err := m.port.ResetOutputBuffer(); err != nil {
		errorLog.Printf("[modem] %s:unable to reset output buffer: %s\n", m.tag, err.Error())
	}
}
