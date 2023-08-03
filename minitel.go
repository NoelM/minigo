package minigo

import (
	"context"
	"fmt"
	"time"

	"nhooyr.io/websocket"
)

type AckType uint

const (
	NoAck = iota
	AckRouleau
	AckPage
)

type Minitel struct {
	RecvKey chan uint
	Quit    chan bool

	conn    *websocket.Conn
	ctx     context.Context
	ackType AckType

	terminalByte       byte
	vitesseByte        byte
	fonctionnementByte byte
	protocoleByte      byte
}

func NewMinitel(conn *websocket.Conn, ctx context.Context) *Minitel {
	return &Minitel{
		conn:    conn,
		ctx:     ctx,
		RecvKey: make(chan uint),
		Quit:    make(chan bool),
	}
}

func (m *Minitel) ContextError() error {
	return m.ctx.Err()
}

func (m *Minitel) ackChecker(keyBuffer []byte) (err error) {
	switch keyBuffer[2] {
	case Terminal:
		m.terminalByte = keyBuffer[3]
	case Fonctionnement:
		m.fonctionnementByte = keyBuffer[3]
	case Vitesse:
		m.vitesseByte = keyBuffer[3]
	case Protocole:
		m.protocoleByte = keyBuffer[3]
	default:
		fmt.Printf("not handled response byte: %x\n", keyBuffer[3])
		return
	}

	ok := false
	switch m.ackType {
	case AckRouleau:
		ok = BitReadAt(m.fonctionnementByte, 6)
	case AckPage:
		ok = !BitReadAt(m.fonctionnementByte, 6)
	default:
		fmt.Printf("not handled AckType: %d\n", m.ackType)
		return
	}

	if !ok {
		err = fmt.Errorf("not verified for acknowledgment: %d", m.ackType)
	} else {
		fmt.Printf("verified acknowledgement for: %d\n", m.ackType)
	}

	m.ackType = NoAck
	return
}

func (m *Minitel) Listen() {
	fullRead := true
	var keyBuffer []byte
	var keyValue uint

	var done bool
	var pro bool

	for {
		var err error
		var wsMsg []byte

		if fullRead {
			_, wsMsg, err = m.conn.Read(m.ctx)
			if websocket.CloseStatus(err) == websocket.StatusAbnormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				fmt.Printf("[minigo] %s listen stop: %s\n", time.Now().Format(time.RFC3339), err.Error())

				m.Quit <- true
				return
			}
			fullRead = false
		}

		for id, b := range wsMsg {
			keyBuffer = append(keyBuffer, b)

			done, pro, keyValue, err = ReadKey(keyBuffer)
			if err != nil {
				keyBuffer = []byte{}
			}

			if done {
				if pro {
					err = m.ackChecker(keyBuffer)
					if err != nil {
						fmt.Println(err.Error())
					}
				} else {
					m.RecvKey <- keyValue
				}

				keyBuffer = []byte{}
			}

			if id == len(wsMsg)-1 {
				fullRead = true
			}
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
// MODE PAGE OU ROULEAU
//

func (m *Minitel) RouleauOn() error {
	m.ackType = AckRouleau

	buf, _ := GetProCode(Pro2)
	buf = append(buf, Start, Rouleau)
	return m.Send(buf)
}

func (m *Minitel) RouleauOff() error {
	m.ackType = AckPage

	buf, _ := GetProCode(Pro2)
	buf = append(buf, Stop, Rouleau)
	return m.Send(buf)
}
