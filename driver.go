package minigo

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Driver interface {
	popHead() (byte, error)

	Recv() (byte, error)
	Readable() (bool, error)
	Send(msg []byte) (int, error)
}

type TCPDriver struct {
	conn       net.Conn
	timeout    time.Duration
	recvBufLen int
	recvBufPos int
	recvBuf    []byte
}

func NewTCPDriver(conn net.Conn, timeout time.Duration) *TCPDriver {
	return &TCPDriver{
		conn:       conn,
		timeout:    timeout,
		recvBufLen: 0,
		recvBufPos: 0,
		recvBuf:    make([]byte, 1024),
	}
}

func (t *TCPDriver) popHead() (byte, error) {
	if t.recvBufLen == 0 {
		return 0, io.EOF
	}

	if t.recvBufPos < t.recvBufLen {
		b := t.recvBuf[t.recvBufPos]
		t.recvBufPos++
		return b, nil
	} else {
		t.recvBufPos, t.recvBufLen = 0, 0
		return 0, io.EOF
	}
}

func (t *TCPDriver) Recv() (byte, error) {
	if b, err := t.popHead(); err == nil {
		return b, nil
	} else if err != io.EOF {
		return 0, fmt.Errorf("unable to popHead: %w", err)
	}

	var err error
	if t.timeout > 0 {
		if err = t.conn.SetReadDeadline(time.Now().Add(t.timeout)); err != nil {
			return 0, fmt.Errorf("unable to set timeout: %w", err)
		}
	}

	t.recvBuf = make([]byte, 1024)
	t.recvBufLen, err = t.conn.Read(t.recvBuf)
	if err == io.EOF || err == os.ErrDeadlineExceeded {
		t.recvBufLen, t.recvBufPos = 0, 0
		return 0, err
	} else if err != nil {
		t.recvBufLen, t.recvBufPos = 0, 0
		return 0, fmt.Errorf("unable to read from TCP: %w", err)
	}

	if b, err := t.popHead(); err == nil {
		return b, nil
	} else if err == io.EOF {
		return 0, err
	} else {
		return 0, fmt.Errorf("unable to popHead: %w", err)
	}
}

func (t *TCPDriver) Readable() (bool, error) {
	return true, nil
}

func (t *TCPDriver) Send(buf []byte) (int, error) {
	return t.conn.Write(buf)
}
