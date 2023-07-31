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

		recvChan := make(chan uint)
		go func() {
			var wsMsg []byte
			fullRead := true

			var keyBuffer []byte
			var keyValue uint
			var done bool

			var b byte
			var id int

			for {
				if fullRead {
					_, wsMsg, err = c.Read(ctx)
					if err != nil {
						continue
					}
					fullRead = false
				}

				for id, b = range wsMsg {
					keyBuffer = append(keyBuffer, b)

					done, keyValue, err = minigo.ReadKey(keyBuffer)
					if done || err != nil {
						keyBuffer = []byte{}
					}
					if done {
						recvChan <- keyValue
					}

					if id == len(wsMsg)-1 {
						fullRead = true
					}
				}

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

func chat(c *websocket.Conn, ctx context.Context, recvChan chan uint) {
	for {
		select {
		case msg := <-recvChan:
			fmt.Printf("recieved: %d", msg)
		default:
			continue
		}

		if ctx.Err() != nil {
			return
		}
	}
}
