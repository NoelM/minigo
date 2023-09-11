package main

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"
)

type MessageType int

const (
	MessageIRC MessageType = iota
	MessageTeletel
)

type Message struct {
	Nick string      `json:"nick"`
	Text []byte      `json:"text"`
	Type MessageType `json:"type"`
	Time time.Time   `json:"time"`
}

type MessagesServer struct {
	filePath        string
	file            *os.File
	messages        []Message
	subscribers     map[int]int
	subscriberMaxId int
	mutex           sync.RWMutex
}

func NewMessagesServer(filePath string) *MessagesServer {
	return &MessagesServer{
		filePath: filePath,
	}
}

func (m *MessagesServer) LoadMessages() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var err error
	m.file, err = os.Open(m.filePath)

	if os.IsNotExist(err) {
		if m.file, err = os.Create(m.filePath); err != nil {
			return err
		}
	}

	scanner := bufio.NewScanner(m.file)
	scanner.Split(bufio.ScanLines)

	line := 0
	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal([]byte(scanner.Text()), &msg); err != nil {
			errorLog.Printf("unable to marshal line %d: %s\n", line, err.Error())
			continue
		}

		m.messages = append(m.messages, msg)
	}

	return nil
}
