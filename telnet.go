package minigo

import (
	"net"
	"time"
)

type Telnet struct {
	conn net.Conn

	connected bool
}

func NewTelnet(conn net.Conn) (*Telnet, error) {
	return &Telnet{
		conn: conn,
	}, nil
}

func (t *Telnet) Init() error {
	t.connected = true
	return nil
}

func (t *Telnet) Write(b []byte) error {
	_, err := t.conn.Write(b)

	if err != nil {
		t.connected = false
		return &ConnectorError{code: ClosedConnection, raw: err}
	}

	return nil
}

func (t *Telnet) Read() ([]byte, error) {
	msg := make([]byte, 256)
	t.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	n, err := t.conn.Read(msg)

	if err != nil {
		t.connected = false
		return nil, &ConnectorError{code: ClosedConnection, raw: err}
	}

	return msg[:n], nil
}

func (t *Telnet) Connected() bool {
	return t.connected
}

func (t *Telnet) Disconnect() {
	t.connected = false
}
