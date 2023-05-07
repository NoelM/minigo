package minigo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"nhooyr.io/websocket"
	"os"
	"time"
)

type Driver interface {
	Read(buffer *bytes.Buffer) error
	Write(data []byte) (int, error)
}

type TCPDriver struct {
	conn    net.Conn
	timeout time.Duration
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

type WebSocketDriver struct {
	conn *websocket.Conn
	ctx  context.Context
}

func NewWebSocketDriver(conn *websocket.Conn, ctx context.Context) *WebSocketDriver {
	return &WebSocketDriver{
		conn: conn,
		ctx:  ctx,
	}
}

func (wsd *WebSocketDriver) Read(buffer *bytes.Buffer) error {
	t, msg, err := wsd.conn.Read(wsd.ctx)

	if err != nil {
		return fmt.Errorf("unable to recieve byte: %s", err)
	}
	if t != websocket.MessageBinary {
		return fmt.Errorf("unable to recieve message type: %s", t.String())
	}

	_, err = buffer.Write(msg)
	return err
}

func (wsd *WebSocketDriver) Write(data []byte) (int, error) {
	return len(data), wsd.conn.Write(wsd.ctx, websocket.MessageBinary, data)
}
