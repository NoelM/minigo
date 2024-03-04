package databases

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Message struct {
	Nick string    `json:"nick"`
	Text string    `json:"text"`
	Time time.Time `json:"time"`
}

type MessageDatabase struct {
	filePath    string
	file        *os.File
	messages    []Message
	subscribers map[string]int
	mutex       sync.RWMutex
}

func NewMessageDatabase() *MessageDatabase {
	return &MessageDatabase{
		subscribers: make(map[string]int),
	}
}

func (m *MessageDatabase) LoadMessages(filePath string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.filePath = filePath

	filedb, err := os.OpenFile(m.filePath, os.O_RDONLY|os.O_CREATE, 0755)
	if err != nil {
		errorLog.Printf("unable to get database: %s\n", err.Error())
		return err
	}
	infoLog.Printf("opened database: %s\n", filePath)

	scanner := bufio.NewScanner(filedb)
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
	filedb.Close()

	infoLog.Printf("loaded %d messages from database\n", len(m.messages))

	m.file, err = os.OpenFile(m.filePath, os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		errorLog.Printf("unable to get database: %s\n", err.Error())
		return err
	}
	infoLog.Printf("opened database: %s\n", filePath)

	return nil
}

func (m *MessageDatabase) Subscribe(nick string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.subscribers[nick] = -1

	infoLog.Printf("got a new subscriber with id=%s\n", nick)
}

func (m *MessageDatabase) Resign(nick string) {
	infoLog.Printf("resigned subscriber with id=%s\n", nick)
	delete(m.subscribers, nick)
}

func (m *MessageDatabase) GetMessages(nick string) []Message {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	lastMsg, ok := m.subscribers[nick]
	if !ok {
		warnLog.Printf("unable to find subscriber with id=%s\n", nick)
		return nil
	}

	nbMsg := len(m.messages) - (lastMsg + 1)
	messagesCopy := make([]Message, nbMsg)

	copy(messagesCopy, m.messages[lastMsg+1:])
	m.subscribers[nick] = len(m.messages) - 1

	infoLog.Printf("subscriber id=%s received %d messages\n", nick, nbMsg)
	return messagesCopy
}

func (m *MessageDatabase) HasNewMessage(nick string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	lastMsg, ok := m.subscribers[nick]
	if !ok {
		warnLog.Printf("unable to find subscriber with id=%s\n", nick)
		return false
	}

	return len(m.messages)-(lastMsg+1) > 0
}

func (m *MessageDatabase) PushMessage(msg Message, filterNick bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if filterNick {
		if _, ok := m.subscribers[msg.Nick]; ok {
			// Locally connected user, message already in DB
			return
		}
	}

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
