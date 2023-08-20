package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/NoelM/minigo"
	"nhooyr.io/websocket"
)

var infoLog = log.New(os.Stdout, "[MINICHAT] INFO:", log.Ldate|log.LUTC)
var warnLog = log.New(os.Stdout, "[MINICHAT] WARN:", log.Ldate|log.LUTC)
var errorLog = log.New(os.Stdout, "[MINICHAT] ERROR:", log.Ldate|log.LUTC)

func main() {

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
		if err != nil {
			errorLog.Printf("Unable to open WS connection: %s\n", err.Error())
			return
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")
		infoLog.Printf("New connection from IP=%s\n", r.RemoteAddr)

		c.SetReadLimit(1024)

		ctx, cancel := context.WithTimeout(r.Context(), time.Minute*10)
		defer cancel()

		m := minigo.NewMinitel(c, ctx, false)
		go m.Listen()

		nick := logPage(m)

		ircDvr := NewIrcDriver(string(nick))
		go ircDvr.Loop()

		chatPage(m, ircDvr)
		ircDvr.Quit()

		infoLog.Printf("Minitel session closed for IP=%s nick=%s\n", r.RemoteAddr, nick)
	})

	err := http.ListenAndServe("192.168.1.34:3615", fn)
	log.Fatal(err)
}
