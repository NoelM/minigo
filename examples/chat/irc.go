package main

import (
	"crypto/tls"
	"fmt"
	"time"

	irc "github.com/thoj/go-ircevent"
)

const channel = "#minitel"
const serverssl = "irc.libera.chat:7000"

type IrcDriver struct {
	conn *irc.Connection
	quit bool

	Nick        string
	RecvMessage chan Message
	SendMessage chan Message
}

func NewIrcDriver(nick string) *IrcDriver {
	return &IrcDriver{
		Nick:        nick,
		RecvMessage: make(chan Message),
		SendMessage: make(chan Message),
	}
}

func (i *IrcDriver) Quit() {
	i.conn.Quit()
}

func (i *IrcDriver) sendMessageListner() {
	for !i.quit {
		select {
		case msg := <-i.SendMessage:
			i.conn.Privmsg(channel, msg.Text)
		default:
			continue
		}
	}
}

func (i *IrcDriver) Loop() error {
	i.conn = irc.IRC(i.Nick, "Minitel Client")
	i.conn.UseTLS = true
	i.conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	i.conn.AddCallback("001", func(e *irc.Event) { i.conn.Join(channel) })
	i.conn.AddCallback("366", func(e *irc.Event) {})

	i.conn.AddCallback("PRIVMSG", func(event *irc.Event) {
		i.RecvMessage <- Message{
			Nick: event.Nick,
			Text: event.Message(),
			Type: Message_UTF8,
			Time: time.Now(),
		}
	})

	err := i.conn.Connect(serverssl)
	if err != nil {
		fmt.Printf("Err %s", err)
		return err
	}

	go i.sendMessageListner()
	i.conn.Loop()

	fmt.Printf("[chat] %s disconnected from: irc.libera.chat\n", time.Now().Format(time.RFC3339))
	return nil
}
