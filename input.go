package minigo

type Input struct {
	Value []byte

	m             *Minitel
	refX, refY    int
	width, height int
	prefix        string
	cursor        bool
	active        bool
}

func NewInput(m *Minitel, refX, refY int, width, height int, prefix string, cursor bool) *Input {
	return &Input{
		m:      m,
		refX:   refX,
		refY:   refY,
		width:  width,
		height: height,
		prefix: prefix,
		cursor: cursor,
	}
}

// getAbsoluteXY, return where the cursor should be for a certain lenght of message
func (i *Input) getAbsoluteXY() (x, y int) {
	totalLen := len(i.Value)
	if len(i.prefix) > 0 {
		totalLen += len(i.prefix) + 1
	}
	y = totalLen/i.width + i.refY
	x = totalLen%i.width + i.refX
	return
}

func (i *Input) AppendKey(key byte) {
	command := GetMoveCursorXY(i.getAbsoluteXY())
	command = append(command, key)
	i.m.Send(command)

	i.Value = append(i.Value, key)
}

func (i *Input) Correction() {
	if len(i.Value) == 0 {
		return
	}

	i.Value = i.Value[:len(i.Value)-1]

	command := GetMoveCursorXY(i.getAbsoluteXY())
	command = append(command, GetCleanLineFromCursor()...)
	i.m.Send(command)
}

// Repetition replays the print of the Input section
func (i *Input) Repetition() {
	command := GetMoveCursorXY(i.refX, i.refY)

	if len(i.prefix) > 0 {
		command = append(command, EncodeMessage(i.prefix)...)
		command = append(command, GetMoveCursorRight(1)...)
	}

	if len(i.Value) > 0 {
		command = append(command, i.Value...)
	}

	i.m.Send(command)

	if i.cursor {
		i.m.CursorOn()
	}
}

// ClearScreen only clears the input on the minitel screen
// but it keeps the Value member complete
func (i *Input) ClearScreen() {
	if i.cursor {
		i.m.CursorOff()
	}

	command := []byte{}
	for row := 0; row < i.height; row += 1 {
		command = append(command, GetMoveCursorXY(i.refX, i.refY+row)...)
		// TODO: handle input with a width < rowWidth
		command = append(command, GetCleanLineFromCursor()...)
	}
	i.m.Send(command)
}

// Clear, clears both the screen and the member Value
func (i *Input) Clear() {
	i.Value = []byte{}
	i.ClearScreen()
}

// Activate moves the cursor to its actual position and let it on
func (i *Input) Activate() {
	i.m.MoveCursorXY(i.getAbsoluteXY())
	if i.cursor {
		i.m.CursorOn()
	}

	i.active = true
}

func (i *Input) Deactivate() {
	i.m.CursorOff()

	i.active = false
}
