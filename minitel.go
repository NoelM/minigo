package minigo

import (
	"context"
	"fmt"

	"nhooyr.io/websocket"
)

type Minitel struct {
	conn  *websocket.Conn
	ctx   context.Context
	InKey chan uint
}

func NewMinitel(conn *websocket.Conn, ctx context.Context) Minitel {
	return Minitel{
		conn:  conn,
		ctx:   ctx,
		InKey: make(chan uint),
	}
}

func (m *Minitel) ContextError() error {
	return m.ctx.Err()
}

func (m *Minitel) Listen() {
	fullRead := true
	var keyBuffer []byte
	var keyValue uint
	var done bool

	for {
		var err error
		var wsMsg []byte

		if fullRead {
			_, wsMsg, err = m.conn.Read(m.ctx)
			if err != nil {
				continue
			}
			fullRead = false
		}

		for id, b := range wsMsg {
			keyBuffer = append(keyBuffer, b)

			done, keyValue, err = ReadKey(keyBuffer)
			if done || err != nil {
				keyBuffer = []byte{}
			}
			if done {
				m.InKey <- keyValue
			}

			if id == len(wsMsg)-1 {
				fullRead = true
			}
		}

		if m.ctx.Err() != nil {
			fmt.Printf("minitel listen stop\n")
			return
		}
	}
}

func (m *Minitel) Send(buf []byte) error {
	return m.conn.Write(m.ctx, websocket.MessageBinary, buf)
}

func (m *Minitel) Reset() error {
	buf := GetCleanScreen()
	buf = append(buf, EncodeAttributes(GrandeurNormale, FondNormal, CursorOff)...)
	buf = append(buf, GetMoveCursorXY(1, 2)...)
	return m.Send(buf)
}

//
// CLEANS
//

func (m *Minitel) CleanLine() error {
	return m.Send(GetCleanLine())
}

func (m *Minitel) CleanScreenFromCursor() error {
	return m.Send(GetCleanLineFromCursor())
}

func (m *Minitel) CleanScreenFromXY(x, y int) error {
	buf := GetMoveCursorXY(x, y)
	buf = append(buf, GetCleanScreenFromCursor()...)
	return m.Send(buf)
}

//
// WRITES
//

func (m *Minitel) WriteBytesXY(x, y int, inBuf []byte) error {
	buf := GetMoveCursorXY(x, y)
	buf = append(buf, inBuf...)
	return m.Send(buf)
}

func (m *Minitel) WriteStringXY(x, y int, s string) error {
	buf := GetMoveCursorXY(x, y)
	buf = append(buf, EncodeMessage(s)...)
	return m.Send(buf)
}

func (m *Minitel) WriteAttributes(attributes ...byte) error {
	return m.Send(EncodeAttributes(attributes...))
}

//
// MOVES
//

func (m *Minitel) MoveCursorXY(x, y int) error {
	return m.Send(GetMoveCursorXY(x, y))
}

func (m *Minitel) Return(n int) error {
	return m.Send(GetMoveCursorReturn(n))
}

func (m *Minitel) MoveCursorDown(n int) error {
	return m.Send(GetMoveCursorDown(n))
}

//
// CURSORS
//

func (m *Minitel) CursorOn() error {
	return m.Send(EncodeAttribute(CursorOn))
}

func (m *Minitel) CursorOnXY(x, y int) error {
	buf := GetMoveCursorXY(x, y)
	buf = append(buf, EncodeAttribute(CursorOn)...)
	return m.Send(buf)
}

func (m *Minitel) CursorOff() error {
	return m.Send(EncodeAttribute(CursorOff))
}

//
// PAGE
//

func (m *Minitel) RouleauOn() error {
	buf, _ := GetProCode(2)
	buf = append(buf, Start, Rouleau)
	return m.Send(buf)
}

func (m *Minitel) RouleauOff() error {
	buf, _ := GetProCode(2)
	buf = append(buf, Stop, Rouleau)
	return m.Send(buf)
}
