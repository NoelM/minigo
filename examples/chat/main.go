package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

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

		recvChan := make(chan []byte)
		go func() {
			for {
				_, msg, err := c.Read(ctx)
				if err != nil {
					continue
				}
				recvChan <- msg

				if ctx.Err() != nil {
					return
				}
			}
		}()

		chat(c, ctx, recvChan)
	})

	err := http.ListenAndServe("192.168.1.34:3615", fn)
	log.Fatal(err)
}

func chat(c *websocket.Conn, ctx context.Context, recvChan chan []byte) {
	for {
		select {
		case msg := <-recvChan:
			fmt.Printf("recieved: %s", msg)
		default:
			continue
		}

		if ctx.Err() != nil {
			return
		}
	}
}
