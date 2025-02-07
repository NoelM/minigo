package minigo

import (
	"net"
)

type TCP struct {
	conn net.Conn

	connected bool
}

func NewTCP(conn net.Conn) (*TCP, error) {
	return &TCP{
		conn: conn,
	}, nil
}

func (t *TCP) Init() error {
	t.connected = true
	return nil
}

func (t *TCP) Write(b []byte) error {
	_, err := t.conn.Write(b)

	if err != nil {
		t.connected = false
		return &ConnectorError{code: ClosedConnection, raw: err}
	}

	return nil
}

func (t *TCP) Read() ([]byte, error) {
	msg := make([]byte, 256)
	n, err := t.conn.Read(msg)

	if err != nil {
		t.connected = false
		return nil, &ConnectorError{code: ClosedConnection, raw: err}
	}

	return msg[:n], nil
}

func (t *TCP) Connected() bool {
	return t.connected
}

func (t *TCP) Disconnect() {
	t.conn.Close()
	t.connected = false
}
