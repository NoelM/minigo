package minigo

import (
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

type Minitel struct {
	net   Network
	group *sync.WaitGroup

	defaultCouleur  int32
	defaultGrandeur int32
	currentGrandeur int32
	defaultFond     int32

	ackStack           *AckStack
	terminalByte       byte
	vitesseByte        byte
	fonctionnementByte byte
	protocoleByte      byte
	clavierByte        byte

	source   string
	connLost *prometheus.CounterVec

	In chan int32
}

func NewMinitel(conn Connector, parity bool, source string, connLost *prometheus.CounterVec, group *sync.WaitGroup) *Minitel {
	return &Minitel{
		net:             *NewNetwork(conn, parity, group, source),
		defaultCouleur:  CaractereBlanc,
		defaultGrandeur: GrandeurNormale,
		currentGrandeur: GrandeurNormale,
		defaultFond:     FondNormal,
		ackStack:        NewAckStack(),
		source:          source,
		connLost:        connLost,
		group:           group,
		In:              make(chan int32, 256),
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

func (m *Minitel) saveProtocol(entryBuffer []byte) {
	switch entryBuffer[2] {
	case Terminal:
		m.terminalByte = entryBuffer[3]
	case Fonctionnement:
		m.fonctionnementByte = entryBuffer[3]
	case Vitesse:
		m.vitesseByte = entryBuffer[3]
	case Protocole:
		m.protocoleByte = entryBuffer[3]
	case CodeReceptionClavier:
		m.clavierByte = entryBuffer[3]
	default:
		warnLog.Printf("[%s] ack-checker: not handled response byte: %x\n", m.source, entryBuffer[2])
		return
	}
}

// TODO: this AckChecker, does not ack anything, it only prints a message
func (m *Minitel) ackChecker() {
	var ok bool
	var ack AckType
	nbAck := m.ackStack.Len()

	for i := 0; i < nbAck; i += 1 {
		ack, ok = m.ackStack.Pop()
		if !ok {
			warnLog.Printf("[%s] ack-checker: no remaining ack to check\n", m.source)
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
		case AckClavierEtendu:
			ok = BitReadAt(m.clavierByte, 0)
		case AckClavierStandard:
			ok = !BitReadAt(m.clavierByte, 0)
		default:
			warnLog.Printf("[%s] ack-checker: not handled ackType=%d\n", m.source, ack)
		}

		if !ok {
			errorLog.Printf("[%s] ack-checker: not verified for acknowledgment ackType=%d\n", m.source, ack)
			m.ackStack.Add(ack)

		} else {
			infoLog.Printf("[%s] ack-checker: verified acknowledgement ackType=%d\n", m.source, ack)
		}
	}
}

func (m *Minitel) Serve() {
	var inbyte byte
	var word []byte

	var gotCnxFin bool

	m.net.Serve()

	for m.net.Connected() {

		select {
		case inbyte = <-m.net.In:
		default:
			// No message from the network, we'll wait a bit
			time.Sleep(100 * time.Millisecond)
			continue
		}

		word = append(word, inbyte)

		// Now we read the word
		// done:  is true when the word has a sense
		// pro:   is true if the message is a protocol message
		// entry: is non-zero when 'done' is true
		// err:   stands for words bigger than 4 bytes (uint32)
		done, pro, entry, err := ReadEntryBytes(word)
		if err != nil {
			errorLog.Printf("[%s] listen: unable to read key=%x: %s\n", m.source, word, err.Error())

			word = []byte{}
			continue
		}

		if done {
			// Enters here only if the previous buffer has been full read
			// Now one gets a non-zero 'entry' value
			if pro {
				infoLog.Printf("[%s] listen: received protocol code=%x\n", m.source, word)

				m.saveProtocol(word)
				if !m.ackStack.Empty() {
					m.ackChecker()
				}

			} else {
				m.toApp(entry)

				if entry == ConnexionFin {
					infoLog.Printf("[%s] listen: caught ConnexionFin: quit loop\n", m.source)

					gotCnxFin = true
					break
				}
			}

			// We have read the word properly, let's reset it
			word = []byte{}
		}
	}

	infoLog.Printf("[%s] listen: loop exited\n", m.source)

	if !gotCnxFin {
		// The loop has been exited without a ConnexionFin, one considers a lost connexion issue
		infoLog.Printf("[%s] listen: connection lost: sending ConnexionFin to Page\n", m.source)
		m.connLost.With(prometheus.Labels{"source": m.source}).Inc()

		// The application loop waits for the ConnexionFin signal to quit
		m.toApp(ConnexionFin)
	}

	infoLog.Printf("[%s] listen: end of listen\n", m.source)
	m.group.Done()
}

func (m *Minitel) Send(buf []byte) error {
	m.net.Out <- buf
	return nil
}

func (m *Minitel) toApp(entry int32) {
	m.In <- entry
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
// CLAVIER ETENDU
//

func (m *Minitel) ClavierEtendu() error {
	m.ackStack.Add(AckClavierEtendu)

	return m.Send([]byte{Prog, Start, CodeReceptionClavier, Eten})
}

func (m *Minitel) ClavierStandard() error {
	m.ackStack.Add(AckClavierEtendu)

	return m.Send([]byte{Prog, Stop, CodeReceptionClavier, Eten})
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
