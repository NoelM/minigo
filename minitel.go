package minigo

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/prometheus/client_golang/prometheus"
)

var infoLog = log.New(os.Stdout, "[minigo] info:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var warnLog = log.New(os.Stdout, "[minigo] warn:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var errorLog = log.New(os.Stdout, "[minigo] error:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

const MaxSubPerMinute = 5

type Minitel struct {
	RecvKey chan int32

	conn   Connector
	parity bool
	wg     *sync.WaitGroup

	defaultCouleur  int32
	defaultGrandeur int32
	currentGrandeur int32
	defaultFond     int32

	ackStack           *AckStack
	terminalByte       byte
	vitesseByte        byte
	fonctionnementByte byte
	protocoleByte      byte

	sentBytes  *Stack
	sentBlocks *Stack

	tag      string
	connLost *prometheus.CounterVec

	pce     bool
	pceLock sync.Mutex
}

func NewMinitel(conn Connector, parity bool, tag string, connLost *prometheus.CounterVec, wg *sync.WaitGroup) *Minitel {
	return &Minitel{
		conn:            conn,
		parity:          parity,
		defaultCouleur:  CaractereBlanc,
		defaultGrandeur: GrandeurNormale,
		currentGrandeur: GrandeurNormale,
		defaultFond:     FondNormal,
		ackStack:        NewAckStack(),
		RecvKey:         make(chan int32),
		tag:             tag,
		connLost:        connLost,
		wg:              wg,
		sentBytes:       NewStack(1),
		sentBlocks:      NewStack(16),
	}
}

func (m *Minitel) charWidth() int {
	if m.currentGrandeur == DoubleLargeur || m.currentGrandeur == DoubleGrandeur {
		return 2
	}
	return 1
}

func (m *Minitel) updateGrandeur(attributes ...byte) {
	for _, attr := range attributes {
		if attr >= GrandeurNormale && attr <= DoubleGrandeur {
			m.currentGrandeur = int32(attr)
		}
	}
}

func (m *Minitel) startPCE() (err error) {
	if !m.pceLock.TryLock() {
		return nil
	}
	m.ackStack.Add(AckPCEStart)

	buf, _ := GetProCode(Pro2)
	buf = append(buf, Start, PCE)
	return m.conn.Write(buf)
}

func (m *Minitel) stopPCE() (err error) {
	m.ackStack.Add(AckPCEStop)

	buf, _ := GetProCode(Pro2)
	buf = append(buf, Stop, PCE)
	return m.conn.Write(buf)
}

func (m *Minitel) PCEMessage() {
	m.WriteStatusLine("→ Mauvaise connexion: PCE ON")
}

func (m *Minitel) ackChecker(keyBuffer []byte) (ack AckType, err error) {
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
		warnLog.Printf("[%s] ack-checker: not handled response byte: %x\n", m.tag, keyBuffer[2])
		return
	}

	var ok bool

	ack, ok = m.ackStack.Pop()
	if !ok {
		warnLog.Printf("[%s] ack-checker: no remaining ack to check\n", m.tag)
		return
	}

	switch ack {
	case AckRouleau:
		ok = BitReadAt(m.fonctionnementByte, 1)
	case AckPage:
		ok = !BitReadAt(m.fonctionnementByte, 1)
	case AckMinuscule:
		ok = BitReadAt(m.fonctionnementByte, 3)
	case AckMajuscule:
		ok = !BitReadAt(m.fonctionnementByte, 3)
	case AckPCEStart:
		if ok = BitReadAt(m.fonctionnementByte, 2); ok {
			m.pce = true
			m.pceLock.Unlock()

			m.PCEMessage()
		}
	case AckPCEStop:
		if ok = !BitReadAt(m.fonctionnementByte, 2); ok {
			m.pce = false
		}
	default:
		warnLog.Printf("[%s] ack-checker: not handled ackType=%d\n", m.tag, ack)
		return
	}

	if !ok {
		err = fmt.Errorf("not verified for acknowledgment ackType=%d", ack)
		errorLog.Printf("[%s] ack-checker: %s\n", m.tag, err.Error())
	} else {
		infoLog.Printf("[%s] ack-checker: verified acknowledgement ackType=%d\n", m.tag, ack)
	}

	return
}

func (m *Minitel) Listen() {
	var keyBuffer []byte
	var keyValue int32

	var done bool
	var pro bool
	var nack bool

	var cnxFinRcvd bool

	// Sub is a message for bad lines transmissions
	var cntSub int
	var firstSub time.Time

	for m.IsConnected() {
		var err error
		var inBytes []byte

		inBytes, err = m.conn.Read()
		if err != nil {
			warnLog.Printf("[%s] listen: stop loop: lost connection: %s\n", m.tag, err.Error())
			break
		}
		if len(inBytes) == 0 {
			continue
		}

		var parityErr error
		for _, b := range inBytes {

			if m.parity {
				if b, parityErr = CheckByteParity(b); parityErr != nil {
					errorLog.Printf("[%s] listen: wrong parity ignored key=%x\n", m.tag, b)

					keyBuffer = []byte{}
					continue
				}
			}

			keyBuffer = append(keyBuffer, b)

			done, pro, keyValue, err = ReadKey(keyBuffer)
			if err != nil {
				errorLog.Printf("[%s] listen: unable to read key=%x: %s\n", m.tag, keyBuffer, err.Error())

				keyBuffer = []byte{}
				continue
			}

			if done {
				switch keyValue {
				case Sub:
					if time.Since(firstSub) < time.Minute {
						cntSub += 1
						infoLog.Printf("[%s] listen: recv SUB, first=%.0fs cnt=%d pce=%t\n", m.tag, time.Since(firstSub).Seconds(), cntSub, m.pce)

						if cntSub > MaxSubPerMinute && !m.pce {
							infoLog.Printf("[%s] listen: too many SUB cnt=%d pce=%t: activate PCE\n", m.tag, cntSub, m.pce)
							m.startPCE()
						}

					} else {
						cntSub = 1
						infoLog.Printf("[%s] listen: recv SUB, reset first=%.0fs cnt=%d pce=%t\n", m.tag, time.Since(firstSub).Seconds(), cntSub, m.pce)

						firstSub = time.Now()
					}

					keyBuffer = []byte{}
					continue

				case Nack:
					infoLog.Printf("[%s] listen: recv NACK\n", m.tag)

					nack = true
					m.pceLock.Lock()

					keyBuffer = []byte{}
					continue
				}

				if nack {
					blockId := int(keyValue - 0x40)
					infoLog.Printf("[%s] listen: recv block to repeat val=%x id=%d\n", m.tag, keyValue, blockId)

					m.synSend(blockId)

					m.pceLock.Unlock()
					nack = false

				} else if pro {
					infoLog.Printf("[%s] listen: received protocol code=%x\n", m.tag, keyBuffer)

					_, err = m.ackChecker(keyBuffer)
					if err != nil {
						errorLog.Printf("[%s] listen: unable to acknowledge protocol code=%x: %s\n", m.tag, keyBuffer, err.Error())
					}

				} else {
					m.RecvKey <- keyValue

					if keyValue == ConnexionFin {
						infoLog.Printf("[%s] listen: caught ConnexionFin: quit loop\n", m.tag)

						cnxFinRcvd = true
						break
					}
				}

				keyBuffer = []byte{}
			}
		}

		// The CnxFin only breaks the previous loop
		if cnxFinRcvd {
			break
		}
	}
	infoLog.Printf("[%s] listen: loop exited\n", m.tag)

	if !cnxFinRcvd {
		infoLog.Printf("[%s] listen: connection lost: sending ConnexionFin to Page\n", m.tag)
		m.connLost.With(prometheus.Labels{"source": m.tag}).Inc()

		m.RecvKey <- ConnexionFin
	}

	infoLog.Printf("[%s] listen: end of listen\n", m.tag)
	m.wg.Done()
}

func (m *Minitel) IsConnected() bool {
	return m.conn.Connected()
}

func (m *Minitel) Send(buf []byte) error {
	m.pceLock.Lock()
	defer m.pceLock.Unlock()

	return m.freeSend(buf)
}

func (m *Minitel) synSend(id int) error {
	block := m.sentBlocks.Get(id)
	if block == nil {
		warnLog.Printf("[%s] syn-send: cannot repeat block id=%d\n", m.tag, id)

		block = m.prepareMsg([]byte{0})[0]
		m.sentBlocks.Add(block)
	}

	buf := []byte{Syn, Syn, 0x40 + byte(id)}
	for id, b := range buf {
		buf[id] = GetByteWithParity(b)
	}

	buf = append(buf, block...)
	return m.conn.Write(buf)
}

func (m *Minitel) freeSend(buf []byte) error {
	prepared := m.prepareMsg(buf)

	for _, msg := range prepared {
		if m.pce {
			m.sentBlocks.Add(msg)
		} else {
			m.sentBytes.Add(msg)
		}

		m.conn.Write(msg)
	}
	return nil
}

func (m *Minitel) prepareMsg(msg []byte) (prepared [][]byte) {
	if m.parity {
		for id, b := range msg {
			msg[id] = GetByteWithParity(b)
		}
	}

	if m.pce {
		for pos := 0; pos < len(msg); pos += 15 {
			prepared = append(prepared, ComputePCEBlock(msg[pos:]))
		}
		return prepared

	} else {
		return append(prepared, msg)
	}
}

func (m *Minitel) Reset() error {
	buf := GetCleanScreen()
	buf = append(buf, EncodeAttributes(GrandeurNormale, FondNormal, CursorOff)...)
	buf = append(buf, GetMoveCursorAt(1, 1)...)
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

func (m *Minitel) CleanScreenFrom(row, col int) error {
	buf := GetMoveCursorAt(row, col)
	buf = append(buf, GetCleanScreenFromCursor()...)
	return m.Send(buf)
}

func (m *Minitel) CleanNRowsFrom(row, col, n int) error {
	buf := GetMoveCursorAt(row, col)
	buf = append(buf, GetCleanNRowsFromCursor(n)...)
	return m.Send(buf)
}

//
// WRITES
//

func (m *Minitel) WriteStatusLine(s string) error {
	buf := []byte{Us, 0x40, 0x41}
	buf = append(buf, GetRepeatRune(' ', 34)...)
	buf = append(buf, Us, 0x40, 0x41)
	buf = append(buf, EncodeMessage(s)...)
	buf = append(buf, Us)
	return m.Send(buf)
}

func (m *Minitel) WriteBytesAt(lineId, colId int, inBuf []byte) error {
	buf := GetMoveCursorAt(lineId, colId)
	buf = append(buf, inBuf...)
	return m.Send(buf)
}

func (m *Minitel) WriteStringLeft(lineId int, s string) error {
	return m.WriteStringAt(lineId, 1, s)
}

func (m *Minitel) WriteNRunes(r rune, n int) error {
	return m.Send(GetRepeatRune(r, n))
}

func (m *Minitel) WriteStringRight(lineId int, s string) error {
	msgLen := utf8.RuneCountInString(s) * m.charWidth()
	colId := maxInt(ColonnesSimple-msgLen+1, 0)

	return m.WriteStringAt(lineId, colId, s)
}

func (m *Minitel) WriteStringCenter(lineId int, s string) error {
	msgLen := len(s) * m.charWidth()
	colId := maxInt((ColonnesSimple-msgLen)/2+1, 0)

	return m.WriteStringAt(lineId, colId, s)
}

func (m *Minitel) WriteStringAt(lineId, colId int, s string) error {
	buf := GetMoveCursorAt(lineId, colId)
	buf = append(buf, EncodeMessage(s)...)
	return m.Send(buf)
}

func (m *Minitel) WriteStringAtWithAttributes(lineId, colId int, s string, attributes ...byte) error {
	m.WriteAttributes(attributes...)

	buf := GetMoveCursorAt(lineId, colId)
	buf = append(buf, EncodeMessage(s)...)
	m.Send(buf)

	return m.WriteAttributes(byte(m.defaultCouleur), byte(m.defaultFond), byte(m.defaultGrandeur))
}

func (m *Minitel) WriteAttributes(attributes ...byte) error {
	m.updateGrandeur(attributes...)

	return m.Send(EncodeAttributes(attributes...))
}

func (m *Minitel) WriteHelperAt(lineId, colId int, helpText, button string) error {
	m.WriteStringAt(lineId, colId, helpText)

	helpMsgLen := (len(helpText) + 1) * m.charWidth()
	buttonCol := minInt(colId+helpMsgLen, ColonnesSimple)
	return m.WriteStringAtWithAttributes(lineId, buttonCol, button, InversionFond)
}

func (m *Minitel) WriteHelperLeft(lineId int, helpText, button string) error {
	m.WriteStringLeft(lineId, helpText)

	helpMsgLen := (len(helpText) + 2) * m.charWidth()
	buttonCol := minInt(helpMsgLen, ColonnesSimple)
	return m.WriteStringAtWithAttributes(lineId, buttonCol, button, InversionFond)
}

func (m *Minitel) WriteHelperRight(lineId int, helpText, button string) error {
	startCol := ColonnesSimple - m.charWidth()*(len(helpText)+len(button)+1) // free space
	startCol = maxInt(startCol, 0)

	m.WriteStringAt(lineId, startCol, helpText)

	buttonCol := minInt(startCol+(1+len(helpText))*m.charWidth(), ColonnesSimple)
	return m.WriteStringAtWithAttributes(lineId, buttonCol, button, InversionFond)
}

//
// MOVES
//

func (m *Minitel) MoveCursorAt(lineId, colId int) error {
	return m.Send(GetMoveCursorAt(lineId, colId))
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

func (m *Minitel) CursorOnXY(col, row int) error {
	buf := GetMoveCursorAt(row, col)
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
	m.ackStack.Add(AckRouleau)

	buf, _ := GetProCode(Pro2)
	buf = append(buf, Start, Rouleau)
	return m.Send(buf)
}

func (m *Minitel) RouleauOff() error {
	m.ackStack.Add(AckPage)

	buf, _ := GetProCode(Pro2)
	buf = append(buf, Stop, Rouleau)
	return m.Send(buf)
}

//
// MINUSCULES
//

func (m *Minitel) MinusculeOn() error {
	m.ackStack.Add(AckMinuscule)

	buf, _ := GetProCode(Pro2)
	buf = append(buf, Start, Minuscules)
	return m.Send(buf)
}

func (m *Minitel) MinusculeOff() error {
	m.ackStack.Add(AckMajuscule)

	buf, _ := GetProCode(Pro2)
	buf = append(buf, Stop, Minuscules)
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

//
// G0, G1, G2
//

func (m *Minitel) ModeG0() error {
	return m.Send([]byte{Si})
}

func (m *Minitel) ModeG1() error {
	return m.Send([]byte{So})
}

func (m *Minitel) ModeG2() error {
	return m.Send([]byte{Ss2})
}
