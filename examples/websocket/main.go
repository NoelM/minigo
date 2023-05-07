package main

import (
	"context"
	"fmt"
	"github.com/NoelM/minigo"
	"log"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

func main() {
	// This handler demonstrates how to correctly handle a write only WebSocket connection.
	// i.e you only expect to write messages and do not expect to read any messages.
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
		if err != nil {
			log.Println(err)
			return
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")

		ctx, cancel := context.WithTimeout(r.Context(), time.Minute*10)
		defer cancel()

		ctx = c.CloseRead(ctx)

		wsd := minigo.NewWebSocketDriver(c, ctx)
		mini := minigo.NewMinitel(wsd)

		err = mini.PrintMessage("BONJOUR")
		if err != nil {
			fmt.Println(err)
		}
	})

	err := http.ListenAndServe("192.168.1.27:3615", fn)
	log.Fatal(err)
}
