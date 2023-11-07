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
	command := GetMoveCursorAt(i.getCursorPos())
	command = append(command, EncodeRune(r)...)
	i.m.Send(command)

	i.Value = utf8.AppendRune(i.Value, r)
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

	command := GetMoveCursorAt(i.getCursorPos())
	if i.dots {
		command = append(command, EncodeMessage(".")...)
	} else {
		command = append(command, EncodeMessage(" ")...)
	}
	command = append(command, GetMoveCursorAt(i.getCursorPos())...)
	i.m.Send(command)
}

// UnHide reveals the input on screen
func (i *Input) UnHide() {
	command := GetMoveCursorAt(i.refRow, i.refCol)

	if len(i.Value) > 0 {
		command = append(command, i.Value...)
	}

	rowAbs, colAbs := i.getCursorPos()
	for row := rowAbs; row < i.refRow+i.height; row += 1 {
		command = append(command, GetMoveCursorAt(row, i.refCol)...)

		paddingZone := i.width
		if row == rowAbs {
			paddingZone = (i.refCol + i.width) - colAbs
		}

		if i.dots {
			command = append(command, GetRepeatRune('.', paddingZone-1)...)
		} else {
			command = append(command, GetRepeatRune(' ', paddingZone-1)...)
		}
	}

	i.m.Send(command)
}

// Hide clears the input on the Minitel screen,
// but it keeps the Value member complete
func (i *Input) Hide() {
	i.m.CursorOff()

	command := []byte{}
	for row := 0; row < i.height; row += 1 {
		command = append(command, GetMoveCursorAt(i.refRow+row, i.refCol)...)
		command = append(command, GetCleanNItemsFromCursor(i.width)...)
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
	i.m.MoveCursorAt(i.getCursorPos())
	i.m.CursorOn()
}

func (i *Input) Deactivate() {
	i.m.CursorOff()
}
