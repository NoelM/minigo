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

		recvKey := make(chan uint)
		go listenKeys(c, ctx, recvKey)

		chat(c, ctx, recvKey)
	})

	err := http.ListenAndServe("192.168.1.34:3615", fn)
	log.Fatal(err)
}

func listenKeys(c *websocket.Conn, ctx context.Context, recvChan chan uint) {
	fullRead := true
	var keyBuffer []byte
	var keyValue uint
	var done bool

	for {
		var err error
		var wsMsg []byte

		if fullRead {
			_, wsMsg, err = c.Read(ctx)
			if err != nil {
				continue
			}
			fullRead = false
		}

		for id, b := range wsMsg {
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
}

func chat(c *websocket.Conn, ctx context.Context, recvKey chan uint) {
	userInput := []byte{}

	for {
		select {
		case key := <-recvKey:
			if key == minigo.Envoi {
				sendMessage(c, ctx, userInput)
				userInput = []byte{}

			} else if minigo.IsUintAValidChar(key) {
				updateMessageInput(c, ctx, len(userInput), byte(key))
				userInput = append(userInput, byte(key))
			} else {
				fmt.Printf("key: %d not supported", key)
			}
		default:
			continue
		}

		if ctx.Err() != nil {
			return
		}
	}
}

func sendMessage(c *websocket.Conn, ctx context.Context, msg []byte) {
	buf := minigo.GetMoveCursorXY(1, 20)
	buf = append(buf, minigo.GetCleanScreenFromCursor()...)
	buf = append(buf, minigo.GetMoveCursorXY(0, 1)...)
	buf = append(buf, msg...)
	c.Write(ctx, websocket.MessageBinary, buf)
}

func updateMessageInput(c *websocket.Conn, ctx context.Context, len int, key byte) {
	y := len / 40
	x := len % 40

	buf := minigo.GetMoveCursorXY(x+1, y+20)
	buf = append(buf, key)
	c.Write(ctx, websocket.MessageBinary, buf)
}
