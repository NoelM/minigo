package main

import "sync"

type MessageType int64

const (
	Message_UTF8 MessageType = iota
	Message_Teletel
)

type Message struct {
	Nick string
	Text string
	Type MessageType
}

type Messages struct {
	List []Message
	Mtx  sync.RWMutex
}

func (m *Messages) AppendTeletelMessage(nick string, text []byte) {
	m.Mtx.Lock()
	defer m.Mtx.Unlock()

	m.List = append(m.List, Message{
		Nick: nick,
		Text: string(text),
		Type: Message_Teletel,
	})
}

func (m *Messages) AppendMessage(nick string, text string) {
	m.Mtx.Lock()
	defer m.Mtx.Unlock()

	m.List = append(m.List, Message{
		Nick: nick,
		Text: text,
		Type: Message_UTF8,
	})
}
