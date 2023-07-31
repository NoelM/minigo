package main

import (
	"crypto/tls"
	"fmt"

	irc "github.com/thoj/go-ircevent"
)

const channel = "#go-eventirc-test"
const serverssl = "irc.freenode.net:7000"

func main() {
	ircnick1 := "blatiblat"
	irccon := irc.IRC(ircnick1, "IRCTestSSL")
	//irccon.VerboseCallbackHandler = true
	//irccon.Debug = true
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(channel) })
	irccon.AddCallback("366", func(e *irc.Event) {})
	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		msg := event.Message()
		nick := event.Nick
		fmt.Printf("%s: %s\n", nick, msg)
	})
	err := irccon.Connect(serverssl)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	irccon.Loop()
}
