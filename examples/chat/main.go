package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/NoelM/minigo"
	"nhooyr.io/websocket"
)

func main() {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
		if err != nil {
			log.Println(err)
			return
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")

		ctx, cancel := context.WithTimeout(r.Context(), time.Minute*10)
		defer cancel()

		fmt.Printf("[chat] %s new connection from: %s\n", time.Now().Format(time.RFC3339), r.RemoteAddr)

		m := minigo.NewMinitel(c, ctx)
		go m.Listen()

		nick := logPage(m)

		ircDvr := NewIrcDriver(string(nick))
		ircDvr.Loop()

		chatPage(m, ircDvr)
		ircDvr.Quit()

		fmt.Printf("close connection from: %s", r.RemoteAddr)
	})

	err := http.ListenAndServe("192.168.1.34:3615", fn)
	log.Fatal(err)
}
