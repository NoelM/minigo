package minigo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

type Minitel struct {
	fontSize   byte
	resolution uint
	driver     Driver
	outBuffer  []byte
	inBuffer   bytes.Buffer
}

func NewMinitel(driver Driver) *Minitel {
	return &Minitel{
		fontSize:   GrandeurNormale,
		resolution: ResolutionSimple,
		driver:     driver,
	}
}

func (m *Minitel) clearOutBuffer() {
	m.outBuffer = []byte{}
}

func (m *Minitel) writeOutBuffer() (int, error) {
	return m.driver.Write(m.outBuffer)
}

func (m *Minitel) writeAndClearOutBuffer() error {
	_, err := m.writeOutBuffer()

	var retry int
	for err != nil && retry < MaxRetry {
		retry++
		_, err = m.writeOutBuffer()
		time.Sleep(10 * time.Millisecond)
	}

	if err == nil {
		m.clearOutBuffer()
	}
	return err
}

func (m *Minitel) WriteBytes(data []byte) {
	m.driver.Write(data)
}

func (m *Minitel) MoveCursorXY(x, y int) error {
	inBound, err := IsPosInBounds(x, y, m.resolution)
	if err != nil {
		return fmt.Errorf("unable to move cursor: %w", err)
	}
	if !inBound {
		return fmt.Errorf("unable to move cursor: values (x=%d,y=%d) out of bound", x, y)
	}

	m.outBuffer = GetMoveCursorXY(m.outBuffer, x, y)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) MoveCursorLeft(n int) error {
	m.outBuffer = GetMoveCursorLeft(m.outBuffer, n)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) MoveCursorRight(n int) error {
	m.outBuffer = GetMoveCursorRight(m.outBuffer, n)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) MoveCursorDown(n int) error {
	m.outBuffer = GetMoveCursorDown(m.outBuffer, n)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) MoveCursorUp(n int) error {
	m.outBuffer = GetMoveCursorUp(m.outBuffer, n)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) MoveCursorReturn(n int) error {
	m.outBuffer = GetMoveCursorReturn(m.outBuffer, n)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) CleanScreen() error {
	m.outBuffer = GetCleanScreen(m.outBuffer)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) CleanScreenFromCursor() error {
	m.outBuffer = GetCleanScreenFromCursor(m.outBuffer)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) CleanScreenToCursor() error {
	m.outBuffer = GetCleanScreenToCursor(m.outBuffer)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) CleanLine() error {
	m.outBuffer = GetCleanLine(m.outBuffer)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) CleanLineFromCursor() error {
	m.outBuffer = GetCleanLineFromCursor(m.outBuffer)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) CleanLineToCursor() error {
	m.outBuffer = GetCleanLineToCursor(m.outBuffer)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) PrintMessage(msg string) error {
	m.outBuffer = GetMessage(m.outBuffer, msg)
	return m.writeAndClearOutBuffer()
}

func (m *Minitel) readByte() (byte, error) {
	if m.inBuffer.Len() == 0 {
		err := m.driver.Read(&m.inBuffer)
		if err != nil {
			return 0, err
		}
	}

	b, err := m.inBuffer.ReadByte()
	if err != nil {
		return 0, err
	}

	// Seems fixed by the JS and Socketel
	//b, err = CheckByteParity(b)
	//if err != nil {
	//	return 0, err
	//}

	return b, nil
}

func (m *Minitel) ReadKey() (uint, error) {

	b, err := m.readByte()
	if err != nil {
		return 0, err
	}
	readBuffer := []byte{b}

	if readBuffer[0] == 0x19 {
		b, err = m.readByte()
		if err != nil {
			return 0, err
		}
		readBuffer = append(readBuffer, b)

		switch readBuffer[1] {
		case 0x23:
			readBuffer = []byte{0xA3}
		case 0x27:
			readBuffer = []byte{0xA7}
		case 0x30:
			readBuffer = []byte{0xB0}
		case 0x31:
			readBuffer = []byte{0xB1}
		case 0x38:
			readBuffer = []byte{0xF7}
		case 0x7B:
			readBuffer = []byte{0xDF}
		}
	} else if readBuffer[0] == 0x13 {
		b, err = m.readByte()
		if err != nil {
			return 0, err
		}
		readBuffer = append(readBuffer, b)
	} else if readBuffer[0] == 0x1B {
		time.Sleep(20 * time.Millisecond)
		b, err = m.readByte()
		if err != nil {
			return 0, err
		}
		readBuffer = append(readBuffer, b)

		if readBuffer[1] == 0x5B {
			b, err = m.readByte()
			if err != nil {
				return 0, err
			}
			readBuffer = append(readBuffer, b)

			if readBuffer[2] == 0x34 || readBuffer[2] == 0x32 {
				b, err = m.readByte()
				if err != nil {
					return 0, err
				}
				readBuffer = append(readBuffer, b)
			}
		}
	}

	switch len(readBuffer) {
	case 1:
		return uint(readBuffer[0]), nil
	case 2:
		return uint(binary.BigEndian.Uint16(readBuffer)), nil
	case 3:
		return uint(binary.BigEndian.Uint32(readBuffer)), nil
	case 4:
		return uint(binary.BigEndian.Uint64(readBuffer)), nil
	default:
		return 0, errors.New("unable to cast inBuffer")
	}
}

func (m *Minitel) IsClosed() bool {
	return m.driver.IsClosed()
}

func (m *Minitel) CursorOn() {
	buf := []byte{}

	buf = GetCursorOn(buf)
	buf = append(buf, Esc, 0x23, 0x20, 0x5F)
	m.WriteBytes(buf)
}
