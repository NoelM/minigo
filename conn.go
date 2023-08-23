package minigo

import (
	"fmt"
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
	InvalidDefinition = iota
	InvalidInit
	ClosedConnection
	Unreachable
)

type ConnectorError struct {
	code ConnectorErrorCode
	raw  error
}

func (ce *ConnectorError) Code() ConnectorErrorCode {
	return ce.code
}

func (ce *ConnectorError) Raw() error {
	return ce.raw
}

func (ce *ConnectorError) Error() string {
	return fmt.Sprintf("connector error id=%d: %s", ce.code, ce.raw.Error())
}
