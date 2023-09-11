package minigo

import (
	"fmt"
	"io"
	"log"
	"os"
)

var infoLog = log.New(os.Stdout, "[minigo] info:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var warnLog = log.New(os.Stdout, "[minigo] warn:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var errorLog = log.New(os.Stdout, "[minigo] error:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

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

	defaultCouleur  uint
	defaultGrandeur uint
	currentGrandeur uint
	defaultFond     uint

	terminalByte       byte
	vitesseByte        byte
	fonctionnementByte byte
	protocoleByte      byte
	parity             bool
}

func NewMinitel(conn Connector, parity bool) *Minitel {
	return &Minitel{
		conn:            conn,
		parity:          parity,
		RecvKey:         make(chan uint),
		Quit:            make(chan bool),
		defaultCouleur:  CaractereBlanc,
		defaultGrandeur: GrandeurNormale,
		currentGrandeur: GrandeurNormale,
		defaultFond:     FondNormal,
	}
}

func (m *Minitel) charWidth() int {
	if m.currentGrandeur == DoubleLargeur || m.currentGrandeur == DoubleGrandeur {
		return 2
	}
	return 1
}

func (m *Minitel) updateGrandeur(attrbutes ...byte) {
	for _, attr := range attrbutes {
		if attr >= GrandeurNormale && attr <= DoubleGrandeur {
			m.currentGrandeur = uint(attr)
		}
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

	for m.conn.Connected() {
		var err error
		var inBytes []byte

		if fullRead {
			inBytes, err = m.conn.Read()
			if err != nil {
				warnLog.Printf("stop minitel listen: closed connection: %s\n", err.Error())
				m.Quit <- true
			}

			fullRead = false
		}

		var parityErr error
		for id, b := range inBytes {
			if m.parity {
				b, parityErr = CheckByteParity(b)
				if parityErr != nil {
					warnLog.Printf("key=%x ignored: wrong parity\n", b)
					continue
				}
			}

			keyBuffer = append(keyBuffer, b)

			done, pro, keyValue, err = ReadKey(keyBuffer)
			if err != nil {
				errorLog.Printf("Unable to read key=%x: %s\n", keyBuffer, err.Error())
				keyBuffer = []byte{}
			}

			if done {
				if pro {
					infoLog.Printf("Recieved procode=%x\n", keyBuffer)
					err = m.ackChecker(keyBuffer)
					if err != nil {
						errorLog.Printf("Unable to acknowledge procode=%x: %s\n", keyBuffer, err.Error())
					}
				} else {
					m.RecvKey <- keyValue
				}

				keyBuffer = []byte{}
			}

			if id == len(inBytes)-1 {
				fullRead = true
			}
		}
	}

	warnLog.Printf("stop minitel listen: closed connection\n")
	m.Quit <- true
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
	buf = append(buf, GetMoveCursorAt(1, 2)...)
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
	buf := GetMoveCursorAt(x, y)
	buf = append(buf, GetCleanScreenFromCursor()...)
	return m.Send(buf)
}

//
// WRITES
//

func (m *Minitel) WriteBytesAt(colId, lineId int, inBuf []byte) error {
	buf := GetMoveCursorAt(colId, lineId)
	buf = append(buf, inBuf...)
	return m.Send(buf)
}

func (m *Minitel) WriteStringLeft(lineId int, s string) error {
	return m.WriteStringAt(1, lineId, s)
}

func (m *Minitel) WriteStringRight(lineId int, s string) error {
	msgLen := len(s) * m.charWidth()
	colId := maxInt(ColonnesSimple-msgLen, 0)

	return m.WriteStringAt(colId, lineId, s)
}

func (m *Minitel) WriteStringCenter(lineId int, s string) error {
	msgLen := len(s) * m.charWidth()
	colId := maxInt((ColonnesSimple-msgLen)/2, 0)

	return m.WriteStringAt(colId, lineId, s)
}

func (m *Minitel) WriteStringAt(colId, lineId int, s string) error {
	buf := GetMoveCursorAt(colId, lineId)
	buf = append(buf, EncodeMessage(s)...)
	return m.Send(buf)
}

func (m *Minitel) WriteStringAtWithAttributes(colId, lineId int, s string, attributes ...byte) error {
	m.WriteAttributes(attributes...)

	buf := GetMoveCursorAt(colId, lineId)
	buf = append(buf, EncodeMessage(s)...)
	m.Send(buf)

	return m.WriteAttributes(byte(m.defaultCouleur), byte(m.defaultFond), byte(m.defaultGrandeur))
}

func (m *Minitel) WriteAttributes(attributes ...byte) error {
	m.updateGrandeur(attributes...)

	return m.Send(EncodeAttributes(attributes...))
}

func (m *Minitel) WriteHelperAt(colId, lineId int, helpText, button string) error {
	m.WriteStringAt(colId, lineId, helpText)

	helpMsgLen := (len(helpText) + 1) * m.charWidth()
	buttonCol := minInt(colId+helpMsgLen, ColonnesSimple)
	return m.WriteStringAtWithAttributes(buttonCol, lineId, button, InversionFond)
}

func (m *Minitel) WriteHelperLeft(lineId int, helpText, button string) error {
	m.WriteStringLeft(lineId, helpText)

	helpMsgLen := (len(helpText) + 2) * m.charWidth()
	buttonCol := minInt(helpMsgLen, ColonnesSimple)
	return m.WriteStringAtWithAttributes(buttonCol, lineId, button, InversionFond)
}

func (m *Minitel) WriteHelperRight(lineId int, helpText, button string) error {
	startCol := ColonnesSimple - m.charWidth()*(len(helpText)+len(button)-2) // free space
	startCol = maxInt(startCol, 0)

	m.WriteStringAt(startCol, lineId, helpText)

	buttonCol := minInt(startCol+(1+len(helpText))*m.charWidth(), ColonnesSimple)
	return m.WriteStringAtWithAttributes(buttonCol, lineId, button, InversionFond)
}

//
// MOVES
//

func (m *Minitel) MoveCursorAt(colId, lineId int) error {
	return m.Send(GetMoveCursorAt(colId, lineId))
}

func (m *Minitel) Return(n int) error {
	return m.Send(GetMoveCursorReturn(n))
}

func (m *Minitel) MoveCursorDown(n int) error {
	return m.Send(GetMoveCursorDown(n))
}

func (m *Minitel) MoveCursorRight(n int) error {
	return m.Send(GetMoveCursorRight(n))
}

func (m *Minitel) MoveCursorLeft(n int) error {
	return m.Send(GetMoveCursorLeft(n))
}

func (m *Minitel) MoveCursorUp(n int) error {
	return m.Send(GetMoveCursorUp(n))
}

//
// CURSORS
//

func (m *Minitel) CursorOn() error {
	return m.Send(EncodeAttribute(CursorOn))
}

func (m *Minitel) CursorOnXY(x, y int) error {
	buf := GetMoveCursorAt(x, y)
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

//
// VDT FORMAT
//

func (m *Minitel) SendVDT(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	vdt, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return m.Send(vdt)
}
