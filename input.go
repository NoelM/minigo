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

// getCursorPos returns the absolute position of the cursor
func (i *Input) getCursorPos() (row, col int) {
	totalLen := utf8.RuneCount(i.Value)
	if totalLen == i.height*i.width {
		totalLen -= 1 // do not move the cursor to the next pos
	}

	row = totalLen/i.width + i.refRow
	col = totalLen%i.width + i.refCol
	return
}

// Init displays the input empty
func (i *Input) Init() {
	i.UnHide()
}

// AppendKey appends a new Rune to the Value array
func (i *Input) AppendKey(r rune) {
	if utf8.RuneCount(i.Value) == i.width*i.height {
		i.m.Bell()
		return
	}

	row, col := i.getCursorPos()
	command := MoveAt(row, col, i.m.supportCSI)
	command = append(command, EncodeRune(r)...)
	i.m.Send(command)

	i.Value = utf8.AppendRune(i.Value, r)

	if utf8.RuneCount(i.Value) == i.width*i.height {
		i.m.MoveAt(i.getCursorPos())
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
	i.Value = i.Value[:len(i.Value)-shift]

	row, col := i.getCursorPos()
	command := MoveAt(row, col, i.m.supportCSI)
	if i.dots {
		command = append(command, EncodeString(".")...)
	} else {
		command = append(command, EncodeString(" ")...)
	}

	row, col = i.getCursorPos()
	command = append(command, MoveAt(row, col, i.m.supportCSI)...)
	i.m.Send(command)
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

	command := []byte{}
	for row := 0; row < i.height; row += 1 {
		command = append(command, MoveAt(i.refRow+row, i.refCol, i.m.supportCSI)...)
		command = append(command, CleanNItemsFromCursor(i.width)...)
	}
	i.m.Send(command)
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
