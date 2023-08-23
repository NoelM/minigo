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

var infoLog = log.New(os.Stdout, "[notel] info:", log.Ldate|log.LUTC)
var warnLog = log.New(os.Stdout, "[notel] warn:", log.Ldate|log.LUTC)
var errorLog = log.New(os.Stdout, "[notel] error:", log.Ldate|log.LUTC)

func main() {

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
		if err != nil {
			errorLog.Printf("unable to open websocket connection: %s\n", err.Error())
			return
		}

		defer conn.Close(websocket.StatusInternalError, "websocket internal error, quitting")
		infoLog.Printf("new connection from IP=%s\n", r.RemoteAddr)

		conn.SetReadLimit(1024)

		ctx, cancel := context.WithTimeout(r.Context(), time.Minute*10)
		defer cancel()

		ws, _ := minigo.NewWebsocket(conn, ctx)
		_ = ws.Init()

		m := minigo.NewMinitel(ws, false)
		go m.Listen()

		var id int
		for id >= sommaireId {
			switch id {
			case sommaireId:
				id = PageSommaire(m)
			case ircId:
				id = ServiceMiniChat(m)
			default:
				continue
			}
		}

		infoLog.Printf("Minitel session closed for IP=%s\n", r.RemoteAddr)
	})

	err := http.ListenAndServe("192.168.1.34:3615", fn)
	log.Fatal(err)
}
