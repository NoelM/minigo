package main

import (
	"crypto/tls"
	"fmt"

	irc "github.com/thoj/go-ircevent"
)

const channel = "#minitel"
const serverssl = "irc.libera.chat:7000"

func startIRC(envoi chan []byte, messageList *Messages) {
	ircnick1 := "minitel"
	irccon := irc.IRC(ircnick1, "IRCTestSSL")
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(channel) })
	irccon.AddCallback("366", func(e *irc.Event) {})

	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		msg := event.Message()
		nick := event.Nick

		messageList.AppendMessage(nick, msg)
		fmt.Printf("%s: %s\n", nick, msg)
	})

	go func() {
		for {
			select {
			case msg := <-envoi:
				irccon.Privmsg(channel, string(msg))
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
}
