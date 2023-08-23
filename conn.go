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
