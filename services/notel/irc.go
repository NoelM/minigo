package main

import (
	"crypto/tls"
	"fmt"
	"time"

	irc "github.com/thoj/go-ircevent"
)

const ircChannel = "#minitel"
const ircServerURL = "irc.libera.chat:7000"

type IrcDriver struct {
	conn *irc.Connection

	Nick        string
	RecvMessage chan Message
}

func NewIrcDriver(nick string) *IrcDriver {
	return &IrcDriver{
		Nick:        nick,
		RecvMessage: make(chan Message),
	}
}

func (i *IrcDriver) Quit() {
	i.conn.Quit()
}

func (i *IrcDriver) SendMessage(msg Message) {
	i.conn.Privmsg(ircChannel, msg.Text)
}

func (i *IrcDriver) Loop() error {
	i.conn = irc.IRC(i.Nick, fmt.Sprintf("%s connected from a Minitel", i.Nick))
	i.conn.UseTLS = true
	i.conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	i.conn.AddCallback("001", func(e *irc.Event) { i.conn.Join(ircChannel) })
	i.conn.AddCallback("366", func(e *irc.Event) {})

	i.conn.AddCallback("PRIVMSG", func(event *irc.Event) {
		MessageDb.PushMessage(Message{
			Nick: event.Nick,
			Text: event.Message(),
			Time: time.Now(),
		})
	})

	err := i.conn.Connect(ircServerURL)
	if err != nil {
		errorLog.Printf("unable to connect to IRC: %s\n", err.Error())
		return err
	}

	i.conn.Loop()

	infoLog.Println("disconnected from: irc.libera.chat")
	return nil
}
