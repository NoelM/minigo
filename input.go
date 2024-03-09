package minigo

import "unicode/utf8"

type Input struct {
	Value []byte

	m      *Minitel
	refCol int
	refRow int
	width  int
	height int
	dots   bool
}

func NewInput(m *Minitel, refRow, refCol int, width, height int, dots bool) *Input {
	return &Input{
		m:      m,
		refRow: refRow,
		width:  width,
		refCol: refCol,
		height: height,
		dots:   dots,
	}
}

func NewInputWithValue(m *Minitel, value string, refRow, refCol int, width, height int, dots bool) *Input {
	return &Input{
		m:      m,
		Value:  []byte(value),
		refRow: refRow,
		width:  width,
		refCol: refCol,
		height: height,
		dots:   dots,
	}
}

// getCursorPos returns the absolute position of the cursor
func (i *Input) getCursorPos() (row, col int) {
	len := i.Len()
	if len == i.height*i.width {
		len -= 1 // do not move the cursor to the next pos
	}

	row = len/i.width + i.refRow
	col = len%i.width + i.refCol
	return
}

func (i *Input) Len() int {
	return utf8.RuneCount(i.Value)
}

func (i *Input) isReturn() bool {
	return i.Len() > 0 && i.Len()%i.width == 0
}

func (i *Input) isFull() bool {
	return i.Len() == i.width*i.height
}

// Init displays the input empty
func (i *Input) Init() {
	i.UnHide()
}

// AppendKey appends a new Rune to the Value array
func (i *Input) AppendKey(r rune) {
	if i.isFull() {
		i.m.Bell()
		return
	}

	command := EncodeRune(r)
	i.m.Send(command)

	i.Value = utf8.AppendRune(i.Value, r)

	if i.isFull() {
		i.m.MoveLeft(1)

	} else if i.isReturn() {
		i.m.Return(1)
		i.m.MoveLeft(i.refCol)
	}
}

// Correction removes the last key, on screen and within Value
func (i *Input) Correction() {
	if utf8.RuneCount(i.Value) == 0 {
		return
	}

	r, shift := utf8.DecodeLastRune(i.Value)
	if r == utf8.RuneError {
		return
	}

	if !i.isFull() {
		i.m.MoveLeft(1)
	}
	i.Value = i.Value[:len(i.Value)-shift]

	if i.dots {
		i.m.WriteString(".")
	} else {
		i.m.WriteString(" ")
	}
	i.m.MoveLeft(1)

	if i.isReturn() {
		i.m.MoveUp(1)
		i.m.MoveRight(i.width)
	}
}

// UnHide reveals the input on screen
func (i *Input) UnHide() {
	command := []byte{}

	for row := i.refRow; row < i.refRow+i.height; row += 1 {
		command = append(command, MoveAt(row, i.refCol, i.m.supportCSI)...)

		if i.dots {
			command = append(command, RepeatRune('.', i.width-1)...)
		} else {
			command = append(command, RepeatRune(' ', i.width-1)...)
		}
	}
	command = append(command, MoveAt(i.refRow, i.refCol, i.m.supportCSI)...)

	if len(i.Value) > 0 {
		command = append(command, EncodeBytes(i.Value)...)
	}

	i.m.Send(command)
}

// Hide clears the input on the Minitel screen,
// but it keeps the Value member complete
func (i *Input) Hide() {
	i.m.CursorOff()
	i.m.MoveAt(i.refRow, i.refCol)

	for row := 0; row < i.height; row += 1 {
		i.m.WriteRepeat(Sp, i.width)

		if i.refCol+i.width < 39 {
			i.m.Return(1)
			i.m.MoveRight(i.refCol)
		} else {
			i.m.MoveLineStart()
			i.m.MoveRight(i.refCol)
		}
	}
}

// Reset clears both the input on screen and Value
func (i *Input) Reset() {
	i.Value = []byte{}
	i.UnHide()
}

// Activate moves the cursor to its actual position and let it on
func (i *Input) Activate() {
	i.m.MoveAt(i.getCursorPos())
	i.m.CursorOn()
}

func (i *Input) Deactivate() {
	i.m.CursorOff()
}
