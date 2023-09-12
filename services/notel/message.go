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
	Text string      `json:"text"`
	Type MessageType `json:"type"`
	Time time.Time   `json:"time"`
}

type MessageDatabase struct {
	filePath        string
	file            *os.File
	messages        []Message
	subscribers     map[int]int
	subscriberMaxId int
	mutex           sync.RWMutex
}

func (m *MessageDatabase) LoadMessages(filePath string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.filePath = filePath

	var err error
	m.file, err = os.Open(m.filePath)

	if os.IsNotExist(err) {
		if m.file, err = os.Create(m.filePath); err != nil {
			errorLog.Printf("unable to create database: %s\n", err.Error())
			return err
		}
		infoLog.Printf("created database: %s\n", filePath)
	} else if err != nil {
		errorLog.Printf("unable to get database stats: %s\n", err.Error())
		return err
	} else {
		infoLog.Printf("opened database: %s\n", filePath)
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
	infoLog.Printf("loaded %d messages from database\n", len(m.messages))

	return nil
}

func (m *MessageDatabase) Subscribe() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.subscriberMaxId += 1
	m.subscribers[m.subscriberMaxId] = 0

	infoLog.Printf("got a new subscriber with id=%d\n", m.subscriberMaxId)
	return m.subscriberMaxId
}

func (m *MessageDatabase) Resign(subscriberId int) {
	infoLog.Printf("resigned subscriber with id=%d\n", m.subscriberMaxId)
	delete(m.subscribers, subscriberId)
}

func (m *MessageDatabase) GetMessages(subscriberId int) []Message {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	lastMsg, ok := m.subscribers[subscriberId]
	if !ok {
		warnLog.Printf("unable to find subscriber with id=%d\n", subscriberId)
		return nil
	}

	nbMsg := len(m.messages) - (lastMsg + 1)
	messagesCopy := make([]Message, nbMsg)

	copy(messagesCopy, m.messages[lastMsg:])
	m.subscribers[subscriberId] = len(m.messages) - 1

	infoLog.Printf("subscriber id=%d recieved %d messages\n", subscriberId, nbMsg)
	return messagesCopy
}

func (m *MessageDatabase) PushMessage(msg Message) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.messages = append(m.messages, msg)

	buf, err := json.Marshal(msg)
	if err != nil {
		errorLog.Printf("unable to marshal message: %s\n", err.Error())
	}
	buf = append(buf, '\n')

	_, err = m.file.Write(buf)
	if err != nil {
		errorLog.Printf("unable to write to database: %s\n", err.Error())
	}

	infoLog.Printf("sucessfully pushed message of length=%d to database\n", len(msg.Text))
}

func (m *MessageDatabase) Quit() {
	m.file.Close()
}
