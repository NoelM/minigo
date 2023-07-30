package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/NoelM/minigo"
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

		demo(c, ctx)
	})

	err := http.ListenAndServe("192.168.1.27:3615", fn)
	log.Fatal(err)
}

func demo(c *websocket.Conn, ctx context.Context) {
	for {
		hello := minigo.EncodeMessage("SALUT SALUT !!!")
		hello = append(hello, minigo.GetMoveCursorReturn(1)...)
		hello = append(hello, minigo.EncodeAttribute(minigo.DoubleGrandeur)...)
		hello = append(minigo.EncodeMessage("C'EST GRAND !"))

		c.Write(ctx, websocket.MessageBinary, hello)

		time.Sleep(10 * time.Second)

		c.Write(ctx, websocket.MessageBinary, minigo.GetCleanScreen())
	}
}
