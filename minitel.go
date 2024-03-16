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

type Minitel struct {
	net   Network
	group *sync.WaitGroup

	defaultCouleur  int32
	defaultGrandeur int32
	currentGrandeur int32
	defaultFond     int32

	supportCSI  bool
	rouleauMode bool

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

func (m *Minitel) NoCSI() {
	m.supportCSI = false
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

func (m *Minitel) SupportCSI() bool {
	return m.supportCSI
}

func (m *Minitel) saveProtocol(entryBuffer []byte) {

	if entryBuffer[1] == Pro3 {
		switch entryBuffer[3] {
		case CodeReceptionClavier:
			m.clavierByte = entryBuffer[4]
		}

	} else if entryBuffer[1] == Pro2 {
		switch entryBuffer[2] {
		case Terminal:
			m.terminalByte = entryBuffer[3]
		case Fonctionnement:
			m.fonctionnementByte = entryBuffer[3]
		case Vitesse:
			m.vitesseByte = entryBuffer[3]
		case Protocole:
			m.protocoleByte = entryBuffer[3]
		default:
			warnLog.Printf("[%s] ack-checker: not handled response byte: %x\n", m.source, entryBuffer[2])
			return
		}
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
			if ok = BitReadAt(m.fonctionnementByte, 1); ok {
				m.rouleauMode = true
			}
		case AckPage:
			if ok = !BitReadAt(m.fonctionnementByte, 1); ok {
				m.rouleauMode = false
			}
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
	buf := CleanScreen()
	buf = append(buf, EncodeAttributes(GrandeurNormale, FondNormal, CursorOff)...)
	buf = append(buf, MoveAt(1, 1, m.supportCSI)...)
	return m.Send(buf)
}

//
// CLEANS
//

func (m *Minitel) CleanLine() {
	m.Send(CleanLine())
}

func (m *Minitel) CleanScreen() error {
	return m.Send(CleanScreen())
}

func (m *Minitel) CleanScreenFromCursor() error {
	return m.Send(CleanLineFromCursor())
}

func (m *Minitel) CleanScreenFrom(row, col int) error {
	buf := MoveAt(row, col, m.supportCSI)
	buf = append(buf, CleanScreenFromCursor()...)
	return m.Send(buf)
}

func (m *Minitel) CleanNRowsFrom(row, col, n int) error {
	buf := MoveAt(row, col, m.supportCSI)
	buf = append(buf, CleanNRowsFromCursor(n)...)
	return m.Send(buf)
}

//
// WRITES
//

func (m *Minitel) Print(s string) {
	m.Send(EncodeString(s))
}

func (m *Minitel) Printf(format string, a ...any) {
	m.Print(fmt.Sprintf(format, a...))
}

func (m *Minitel) Button(s string, back, front byte) {
	m.Attributes(back, front)
	m.Print(" ")

	m.Print(s)

	m.Attributes(byte(m.defaultFond), byte(m.defaultCouleur))
	m.Print(" ")

}

func (m *Minitel) PrintStatus(s string) error {
	// Enters row 0
	buf := []byte{Us, 0x40, 0x41}
	// Clean curent value
	buf = append(buf, RepeatRune(' ', 35)...)
	// Return carriage to the begining of row 0
	buf = append(buf, Us, 0x40, 0x41)
	// Encode string to VDT format
	buf = append(buf, EncodeString(s)...)
	// Quit the row 0 with LF
	buf = append(buf, Lf)

	return m.Send(buf)
}

func (m *Minitel) PrintBytesAt(lineId, colId int, inBuf []byte) error {
	buf := MoveAt(lineId, colId, m.supportCSI)
	buf = append(buf, inBuf...)
	return m.Send(buf)
}

func (m *Minitel) PrintLeftAt(lineId int, s string) error {
	return m.PrintAt(lineId, 1, s)
}

func (m *Minitel) Repeat(r rune, n int) error {
	return m.Send(RepeatRune(r, n))
}

func (m *Minitel) PrintRightAt(lineId int, s string) error {
	msgLen := utf8.RuneCountInString(s) * m.charWidth()
	colId := maxInt(ColonnesSimple-msgLen+1, 0)

	return m.PrintAt(lineId, colId, s)
}

func (m *Minitel) PrintCenter(s string) {
	msgLen := len(s) * m.charWidth()
	colId := maxInt((ColonnesSimple-msgLen)/2+1, 0)

	m.Right(colId)
	m.Print(s)
}

func (m *Minitel) PrintCenterAt(lineId int, s string) error {
	msgLen := len(s) * m.charWidth()
	colId := maxInt((ColonnesSimple-msgLen)/2+1, 0)

	return m.PrintAt(lineId, colId, s)
}

func (m *Minitel) PrintAt(lineId, colId int, s string) error {
	buf := MoveAt(lineId, colId, m.supportCSI)
	buf = append(buf, EncodeString(s)...)
	return m.Send(buf)
}

func (m *Minitel) PrintAttributesAt(lineId, colId int, s string, attributes ...byte) error {
	m.Attributes(attributes...)

	buf := MoveAt(lineId, colId, m.supportCSI)
	buf = append(buf, EncodeString(s)...)
	m.Send(buf)

	return m.Attributes(byte(m.defaultCouleur), byte(m.defaultFond), byte(m.defaultGrandeur))
}

func (m *Minitel) PrintAttributes(s string, attributes ...byte) error {
	m.Attributes(attributes...)
	m.Send(EncodeString(s))
	return m.Attributes(byte(m.defaultCouleur), byte(m.defaultFond), byte(m.defaultGrandeur))
}

func (m *Minitel) Attributes(attributes ...byte) error {
	m.updateGrandeur(attributes...)

	return m.Send(EncodeAttributes(attributes...))
}

func (m *Minitel) Helper(helpText, button string, back, front byte) {
	m.Print(helpText)

	m.Right(1)

	m.Attributes(back, front)
	m.Print(" ")

	m.Print(button)

	m.Attributes(byte(m.defaultFond), byte(m.defaultCouleur))
	m.Print(" ")
}

func (m *Minitel) HelperAt(row, col int, helpText, button string) {
	m.MoveAt(row, col)
	m.Helper(helpText, button, FondBlanc, CaractereNoir)
}

func (m *Minitel) HelperLeftAt(row int, helpText, button string) {
	m.MoveAt(row, 0)
	m.Helper(helpText, button, FondBlanc, CaractereNoir)
}

func (m *Minitel) HelperRight(helpText, button string, back, front byte) {
	// [HELP TEXT][BLANK=1][SPACE=1][BUTTON TEXT][SPACE=1]
	refCol := ColonnesSimple - m.charWidth()*(utf8.RuneCountInString(helpText)+1+len(button)+2) - 1
	refCol = maxInt(refCol, 0)

	m.LineStart()
	m.Right(refCol)

	m.Helper(helpText, button, back, front)
}

func (m *Minitel) HelperRightAt(row int, helpText, button string) {
	// [HELP TEXT][BLANK=1][SPACE=1][BUTTON TEXT][SPACE=1]
	refCol := ColonnesSimple - m.charWidth()*(utf8.RuneCountInString(helpText)+1+len(button)+2) - 1
	refCol = maxInt(refCol, 0)

	m.MoveAt(row, refCol)
	m.Helper(helpText, button, FondBlanc, CaractereNoir)
}

//
// MOVES
//

func (m *Minitel) MoveAt(lineId, colId int) error {
	return m.Send(MoveAt(lineId, colId, m.supportCSI))
}

// MoveOf moves the cursor relatively from its current position
// * row > 0, moves down
// * col > 0, moves right
func (m *Minitel) MoveOf(lineId, colId int) error {
	return m.Send(MoveOf(lineId, colId, m.supportCSI))
}

func (m *Minitel) Return(n int) error {
	return m.Send(Return(n, m.supportCSI))
}

func (m *Minitel) ReturnCol(n, col int) error {
	return m.Send(ReturnCol(n, col, m.supportCSI))
}

func (m *Minitel) ReturnUp(n int) error {
	return m.Send(ReturnUp(n, m.supportCSI))
}

func (m *Minitel) Down(n int) error {
	return m.Send(MoveDown(n, m.supportCSI))
}

func (m *Minitel) Right(n int) error {
	return m.Send(MoveRight(n, m.supportCSI))
}

func (m *Minitel) Left(n int) error {
	return m.Send(MoveLeft(n, m.supportCSI))
}

func (m *Minitel) Up(n int) error {
	return m.Send(MoveUp(n, m.supportCSI))
}

func (m *Minitel) LineStart() error {
	return m.Send([]byte{Cr})
}

//
// CURSORS
//

func (m *Minitel) CursorOn() error {
	return m.Send(EncodeAttribute(CursorOn))
}

func (m *Minitel) CursorOff() error {
	return m.Send(EncodeAttribute(CursorOff))
}

//
// MODE PAGE OU ROULEAU
//

func (m *Minitel) RouleauOn() error {
	m.ackStack.Add(AckRouleau)

	buf, _ := ProCode(Pro2)
	buf = append(buf, Start, Rouleau)
	return m.Send(buf)
}

func (m *Minitel) RouleauOff() error {
	m.ackStack.Add(AckPage)

	buf, _ := ProCode(Pro2)
	buf = append(buf, Stop, Rouleau)
	return m.Send(buf)
}

//
// CLAVIER ETENDU
//

func (m *Minitel) ClavierEtendu() error {
	m.ackStack.Add(AckClavierEtendu)

	buf, _ := ProCode(Pro3)
	buf = append(buf, Start, CodeReceptionClavier, Eten)
	return m.Send(buf)
}

func (m *Minitel) ClavierStandard() error {
	m.ackStack.Add(AckClavierEtendu)

	buf, _ := ProCode(Pro3)
	buf = append(buf, Stop, CodeReceptionClavier, Eten)
	return m.Send(buf)
}

//
// MINUSCULES
//

func (m *Minitel) MinusculeOn() error {
	m.ackStack.Add(AckMinuscule)

	buf, _ := ProCode(Pro2)
	buf = append(buf, Start, Minuscules)
	return m.Send(buf)
}

func (m *Minitel) MinusculeOff() error {
	m.ackStack.Add(AckMajuscule)

	buf, _ := ProCode(Pro2)
	buf = append(buf, Stop, Minuscules)
	return m.Send(buf)
}

//
// LINES
//

func (m *Minitel) HLine(len int, t LineType) {
	m.Send(HLine(len, t))
}

func (m *Minitel) VLine(len int, t LineType) {
	m.Send(VLine(len, t))
}

func (m *Minitel) HLineAt(row, col, len int, t LineType) {
	m.Send(HLineAt(row, col, len, t, m.supportCSI))
}

func (m *Minitel) VLineAt(row, col, len int, t LineType) {
	m.Send(VLineAt(row, col, len, t, m.supportCSI))
}

func (m *Minitel) RectAt(row, col, width, height int) {
	m.Send(RectangleAt(row, col, width, height, m.supportCSI))
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

func (m *Minitel) SendCAN() error {
	return m.Send([]byte{Can})
}

//
// BELL
//

func (m *Minitel) Bell() {
	m.Send([]byte{Bel})
}
