package minigo

import (
	"time"
)

type Connector interface {
	Init() error

	Write([]byte) error

	Read() ([]byte, error)

	ReadTimeout(time.Duration) ([]byte, error)

	Connected() bool
}

type ConnectorErrorCode int

const (
	ClosedConnection = iota
	InterfaceUnreachable
)

type ConnectorError struct {
	code ConnectorErrorCode
	raw  error
}

func (ce *ConnectorError) Code() ConnectorErrorCode {
	return ce.code
}

func (ce *ConnectorError) Error() error {
	return ce.raw
}
