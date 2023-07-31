package main

import (
	"context"
	"fmt"

	"github.com/NoelM/minigo"
	"nhooyr.io/websocket"
)

func clearInput(c *websocket.Conn, ctx context.Context) {
	buf := minigo.GetMoveCursorXY(1, 20)
	buf = append(buf, minigo.GetCleanScreenFromCursor()...)
	c.Write(ctx, websocket.MessageBinary, buf)
}

func updateScreen(c *websocket.Conn, ctx context.Context, list *Messages) {
	currentLine := 1

	list.Mtx.RLock()
	defer list.Mtx.RUnlock()

	for i := len(list.List) - 1; i >= 0; i -= 1 {
		// 3 because the format is: "nick > text"
		msgLen := len(list.List[i].Nick) + len(list.List[i].Text) + 3

		// 2 because if msgLen < 40, the divide gives 0 and one break another line for readability
		// nick > text
		// <blank>
		// nick > text2
		msgLines := msgLen/40 + 2

		if currentLine+msgLines > 20 {
			break
		}

		buf := minigo.GetMoveCursorXY(0, currentLine)
		buf = append(buf, minigo.EncodeMessage(fmt.Sprintf("%s > ", list.List[i].Nick))...)

		if list.List[i].Type == Message_Teletel {
			buf = append(buf, list.List[i].Text...)
		} else {
			buf = append(buf, minigo.EncodeMessage(list.List[i].Text)...)
		}

		buf = append(buf, minigo.GetCleanLineFromCursor()...)
		buf = append(buf, minigo.GetMoveCursorReturn(1)...)
		buf = append(buf, minigo.GetCleanLine()...)
		c.Write(ctx, websocket.MessageBinary, buf)

		currentLine += msgLines
	}
}

func appendInput(c *websocket.Conn, ctx context.Context, inputLen int, key byte) {
	y := inputLen / 40
	x := inputLen % 40

	buf := minigo.GetMoveCursorXY(x+1, y+20)
	buf = append(buf, key)
	c.Write(ctx, websocket.MessageBinary, buf)
}

func corrInput(c *websocket.Conn, ctx context.Context, inputLen int) {
	y := (inputLen - 1) / 40
	x := (inputLen - 1) % 40

	buf := minigo.GetMoveCursorXY(x+1, y+20)
	buf = append(buf, minigo.GetCleanLineFromCursor()...)
	c.Write(ctx, websocket.MessageBinary, buf)
}

func updateInput(c *websocket.Conn, ctx context.Context, userInput []byte) {
	buf := minigo.GetMoveCursorXY(1, 20)
	buf = append(buf, userInput...)
	c.Write(ctx, websocket.MessageBinary, buf)
}
