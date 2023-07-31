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

		messageList := Messages{}
		go startIRC(&messageList)

		chat(c, ctx, recvKey, &messageList)
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

func chat(c *websocket.Conn, ctx context.Context, recvKey chan uint, messagesList *Messages) {
	userInput := []byte{}

	for {
		select {
		case key := <-recvKey:
			if key == minigo.Envoi {
				messagesList.AppendTeletelMessage("minitel", userInput)

				clearInput(c, ctx)
				updateScreen(c, ctx, messagesList)
				userInput = []byte{}

			} else if key == minigo.Repetition {
				updateScreen(c, ctx, messagesList)
				updateInput(c, ctx, userInput)

			} else if key == minigo.Correction {
				corrInput(c, ctx, len(userInput))
				userInput = userInput[:len(userInput)-2]

			} else if minigo.IsUintAValidChar(key) {
				appendInput(c, ctx, len(userInput), byte(key))
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
