package databases

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/NoelM/minigo/notel/logs"
)

type Message struct {
	Nick string    `json:"nick"`
	Text string    `json:"text"`
	Time time.Time `json:"time"`
}

// Renamed MessageDatabase to Channel
type Channel struct {
	filePath    string
	file        *os.File
	messages    []Message
	subscribers map[string]int
	mutex       sync.RWMutex
}

// ChatManager handles multiple channels
type ChatManager struct {
	channels map[string]*Channel
	mutex    sync.RWMutex
}

// NewChatManager creates a new chat manager
func NewChatManager() *ChatManager {
	return &ChatManager{
		channels: make(map[string]*Channel),
	}
}

// GetChannel returns an existing channel or creates a new one
func (cm *ChatManager) GetChannel(channelName string) *Channel {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if channel, exists := cm.channels[channelName]; exists {
		return channel
	}

	channel := NewChannel()
	cm.channels[channelName] = channel
	return channel
}

// NewChannel creates a new channel (formerly NewMessageDatabase)
func NewChannel() *Channel {
	return &Channel{
		subscribers: make(map[string]int),
	}
}

func (c *Channel) LoadMessages(filePath string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.filePath = filePath

	filedb, err := os.OpenFile(c.filePath, os.O_RDONLY|os.O_CREATE, 0755)
	if err != nil {
		logs.ErrorLog("unable to get database: %s\n", err.Error())
		return err
	}
	logs.InfoLog("opened database: %s\n", filePath)

	scanner := bufio.NewScanner(filedb)
	scanner.Split(bufio.ScanLines)

	line := 0
	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal([]byte(scanner.Text()), &msg); err != nil {
			logs.ErrorLog("unable to marshal line %d: %s\n", line, err.Error())
			continue
		}

		c.messages = append(c.messages, msg)
	}
	filedb.Close()

	logs.InfoLog("loaded %d messages from database\n", len(c.messages))

	c.file, err = os.OpenFile(c.filePath, os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		logs.ErrorLog("unable to get database: %s\n", err.Error())
		return err
	}
	logs.InfoLog("opened database: %s\n", filePath)

	return nil
}

func (c *Channel) Subscribe(nick string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.subscribers[nick] = -1

	logs.InfoLog("got a new subscriber with id=%s\n", nick)
}

func (c *Channel) Resign(nick string) {
	logs.InfoLog("resigned subscriber with id=%s\n", nick)
	delete(c.subscribers, nick)
}

func (c *Channel) GetMessages(nick string) []Message {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	lastMsg, ok := c.subscribers[nick]
	if !ok {
		logs.WarnLog("unable to find subscriber with id=%s\n", nick)
		return nil
	}

	nbMsg := len(c.messages) - (lastMsg + 1)
	messagesCopy := make([]Message, nbMsg)

	copy(messagesCopy, c.messages[lastMsg+1:])
	c.subscribers[nick] = len(c.messages) - 1

	logs.InfoLog("subscriber id=%s received %d messages\n", nick, nbMsg)
	return messagesCopy
}

func (c *Channel) HasNewMessage(nick string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	lastMsg, ok := c.subscribers[nick]
	if !ok {
		logs.WarnLog("unable to find subscriber with id=%s\n", nick)
		return false
	}

	return len(c.messages)-(lastMsg+1) > 0
}

func (c *Channel) PushMessage(msg Message, filterNick bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if filterNick {
		if _, ok := c.subscribers[msg.Nick]; ok {
			// Locally connected user, message already in DB
			return
		}
	}

	c.messages = append(c.messages, msg)

	buf, err := json.Marshal(msg)
	if err != nil {
		logs.ErrorLog("unable to marshal message: %s\n", err.Error())
	}
	buf = append(buf, '\n')

	_, err = c.file.Write(buf)
	if err != nil {
		logs.ErrorLog("unable to write to database: %s\n", err.Error())
	}

	logs.InfoLog("sucessfully pushed message of length=%d to database\n", len(msg.Text))
}

func (c *Channel) GetConnected() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	connected := make([]string, 0, len(c.subscribers))
	for nick := range c.subscribers {
		connected = append(connected, nick)
	}
	return connected
}

func (c *Channel) Quit() {
	c.file.Close()
}
