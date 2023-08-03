package main

import (
	"crypto/tls"
	"fmt"
	"time"

	irc "github.com/thoj/go-ircevent"
)

const channel = "#minitel"
const serverssl = "irc.libera.chat:7000"

func ircLoop(nick string, quit chan bool, envoi chan []byte, messageList *Messages) {
	irccon := irc.IRC(nick, "IRCTestSSL")
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(channel) })
	irccon.AddCallback("366", func(e *irc.Event) {})

	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		msg := event.Message()
		nick := event.Nick
		messageList.AppendMessage(nick, msg)
	})

	go func() {
		for {
			select {
			case msg := <-envoi:
				irccon.Privmsg(channel, string(msg))
			case <-quit:
				irccon.Quit()
			default:
				continue
			}
		}
	}()

	err := irccon.Connect(serverssl)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	irccon.Loop()

	fmt.Printf("[chat] %s disconnected from: irc.libera.chat\n", time.Now().Format(time.RFC3339))
}
