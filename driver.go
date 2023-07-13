package minigo

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"io"
	"net"
	"os"
	"time"
)

type Driver interface {
	Read(buffer *bytes.Buffer) error
	Write(data []byte) (int, error)
	IsClosed() bool
}

type TCPDriver struct {
	conn    net.Conn
	timeout time.Duration
	close   bool
}

func NewTCPDriver(conn net.Conn, timeout time.Duration) *TCPDriver {
	return &TCPDriver{
		conn:    conn,
		timeout: timeout,
	}
}

func (t *TCPDriver) Read(buffer *bytes.Buffer) error {

	var err error
	if t.timeout > 0 {
		if err = t.conn.SetReadDeadline(time.Now().Add(t.timeout)); err != nil {
			return fmt.Errorf("unable to set timeout: %w", err)
		}
	}

	buf := make([]byte, 1024)
	bufLen, err := t.conn.Read(buf)
	if err == io.EOF || err == os.ErrDeadlineExceeded {
		return err
	} else if err != nil {
		return fmt.Errorf("unable to read from TCP: %w", err)
	}

	_, err = buffer.Write(buf[0:bufLen])
	return err
}

func (t *TCPDriver) SendBytes(buf []byte) (int, error) {
	return t.conn.Write(buf)
}

func (t *TCPDriver) IsClosed() bool {
	return t.close
}

type WebSocketDriver struct {
	conn   net.Conn
	closed bool
}

func NewWebSocketDriver(conn net.Conn) *WebSocketDriver {
	return &WebSocketDriver{
		conn:   conn,
		closed: false,
	}
}

func (wsd *WebSocketDriver) Read(buffer *bytes.Buffer) error {
	msg, _, err := wsutil.ReadClientData(wsd.conn)
	if err != nil && errors.Is(err, io.EOF) {
		wsd.closed = true
	}

	_, err = buffer.Write(msg)
	return err
}

func (wsd *WebSocketDriver) Write(data []byte) (int, error) {
	return len(data), wsutil.WriteServerMessage(wsd.conn, ws.OpBinary, data)
}

func (wsd *WebSocketDriver) IsClosed() bool {
	return wsd.closed
}
