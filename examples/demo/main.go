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

	err := http.ListenAndServe("192.168.1.34:3615", fn)
	log.Fatal(err)
}

func demo(c *websocket.Conn, ctx context.Context) {
	for {
		hello := minigo.EncodeMessage("SALUT SALUT !!!")
		hello = append(hello, minigo.GetMoveCursorReturn(2)...)
		c.Write(ctx, websocket.MessageBinary, hello)
		time.Sleep(1 * time.Second)

		big := minigo.EncodeAttributes(minigo.DoubleGrandeur, minigo.InversionFond)
		big = append(big, minigo.EncodeMessage("C'EST GRAND !")...)
		c.Write(ctx, websocket.MessageBinary, big)
		time.Sleep(3 * time.Second)

		resetScreen := minigo.EncodeAttributes(minigo.GrandeurNormale, minigo.FondNormal)
		resetScreen = append(resetScreen, minigo.GetCleanScreen()...)
		resetScreen = append(resetScreen, minigo.GetMoveCursorXY(0, 1)...)
		c.Write(ctx, websocket.MessageBinary, resetScreen)
	}
}
