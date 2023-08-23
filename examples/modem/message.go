package main

import (
	"time"
)

type MessageType int64

const (
	Message_UTF8 MessageType = iota
	Message_Teletel
)

type Message struct {
	Nick string
	Text string
	Type MessageType
	Time time.Time
}
