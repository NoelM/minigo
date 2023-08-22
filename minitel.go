package minigo

import (
	"fmt"
	"log"
	"os"
)

var infoLog = log.New(os.Stdout, "[minigo] INFO:", log.Ldate|log.LUTC)
var warnLog = log.New(os.Stdout, "[minigo] WARN:", log.Ldate|log.LUTC)
var errorLog = log.New(os.Stdout, "[minigo] ERROR:", log.Ldate|log.LUTC)

type AckType uint

const (
	NoAck = iota
	AckRouleau
	AckPage
)

type Minitel struct {
	RecvKey chan uint
	Quit    chan bool

	conn    Connector
	ackType AckType

	terminalByte       byte
	vitesseByte        byte
	fonctionnementByte byte
	protocoleByte      byte
	parity             bool
}

func NewMinitel(conn Connector, parity bool) *Minitel {
	return &Minitel{
		conn:    conn,
		parity:  parity,
		RecvKey: make(chan uint),
		Quit:    make(chan bool),
	}
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
		err = fmt.Errorf("not verified for acknowledgment ackType=%d", m.ackType)
		errorLog.Panicln(err.Error())
	} else {
		infoLog.Printf("verified acknowledgement ackType=%d\n", m.ackType)
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
		fmt.Printf("restart fullread=%v\n", fullRead)
		var err error
		var wsMsg []byte
		var n int

		if fullRead {
			fmt.Println("wait for key")
			n, wsMsg, err = m.conn.Read()
			fmt.Printf("recv byte=%x len=%d\n", wsMsg, n)
			if err != nil {
				warnLog.Printf("stop minitel listen: closed connection: %s\n", err.Error())
				m.Quit <- true
			}

			fullRead = false
		}

		var parityErr error
		for id, b := range wsMsg[:n] {
			fmt.Printf("loop byte=%x\n", b)
			if m.parity {
				b, parityErr = CheckByteParity(b)
				if parityErr != nil {
					warnLog.Printf("key=%x ignored: wrong parity\n", b)
					continue
				}
			}
			fmt.Println("parity OK")

			keyBuffer = append(keyBuffer, b)

			done, pro, keyValue, err = ReadKey(keyBuffer)
			fmt.Printf("read key done=%v pro=%v key=%d err=%s", done, pro, keyBuffer, err)
			if err != nil {
				errorLog.Printf("Unable to read key=%x: %s\n", keyBuffer, err.Error())
				keyBuffer = []byte{}
			}

			if done {
				fmt.Println("DONE")
				if pro {
					infoLog.Printf("Recieved procode=%x\n", keyBuffer)
					err = m.ackChecker(keyBuffer)
					if err != nil {
						errorLog.Printf("Unable to acknowledge procode=%x: %s\n", keyBuffer, err.Error())
					}
				} else {
					fmt.Println("sent key")
					m.RecvKey <- keyValue
				}

				keyBuffer = []byte{}
				fmt.Println("reset key buffer")
			}

			if id == n-1 {
				fullRead = true
			}
		}
	}
}

func (m *Minitel) Send(buf []byte) error {
	if m.parity {
		for id, b := range buf {
			buf[id] = GetByteWithParity(b)
		}
	}
	return m.conn.Write(buf)
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

func (m *Minitel) CleanScreen() error {
	return m.Send(GetCleanScreen())
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
