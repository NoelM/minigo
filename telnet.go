package minigo

import (
	"github.com/reiver/go-telnet"
)

type Telnet struct {
	ctx  telnet.Context
	read telnet.Reader
	wrt  telnet.Writer

	connected bool
}

func NewTelnet(ctx telnet.Context, w telnet.Writer, r telnet.Reader) (*Telnet, error) {
	return &Telnet{
		ctx:  ctx,
		read: r,
		wrt:  w,
	}, nil
}

func (t *Telnet) Init() error {
	t.connected = true
	return nil
}

func (t *Telnet) Write(b []byte) error {
	_, err := t.wrt.Write(b)

	if err != nil {
		t.connected = false
		return &ConnectorError{code: ClosedConnection, raw: err}
	}

	return nil
}

func (t *Telnet) Read() ([]byte, error) {
	msg := make([]byte, 256)
	n, err := t.read.Read(msg)

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
