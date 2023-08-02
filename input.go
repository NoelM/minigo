package minigo

type Input struct {
	m             *Minitel
	refX, refY    int
	width, height int
	Value         []byte
	pre           string
	cursor        bool
}

func NewInput(m *Minitel, refX, refY int, width, height int, pre string, cursor bool) *Input {
	return &Input{
		m:      m,
		refX:   refX,
		refY:   refY,
		width:  width,
		height: height,
		pre:    pre,
		cursor: cursor,
	}
}

func (i *Input) clearScreen() {
	command := []byte{}

	for row := 0; row < i.height; row += 1 {
		command = GetMoveCursorXY(i.refX, i.refY+row)
		// TODO: handle input with a width < rowWidth
		command = append(command, GetCleanScreenFromCursor()...)
	}

	command = append(command, GetMoveCursorXY(i.refX, i.refY)...)
	command = append(command, EncodeMessage(i.pre)...)
	i.m.Send(command)
}

func (i *Input) getAbsoluteXY(offset int) (x, y int) {
	totalLen := len(i.Value)
	if len(i.pre) > 0 {
		totalLen += len(i.pre) + 1
	}
	y = (totalLen+offset)/i.width + i.refY
	x = (totalLen+offset)%i.width + i.refX
	return
}

func (i *Input) AppendKey(key byte) {
	i.Value = append(i.Value, key)

	x, y := i.getAbsoluteXY(0)

	command := GetMoveCursorXY(x, y)
	command = append(command, key)
	i.m.Send(command)
}

func (i *Input) Correction() {
	if len(i.Value) == 0 {
		return
	}

	i.Value = i.Value[:len(i.Value)-1]
	x, y := i.getAbsoluteXY(0)

	command := GetMoveCursorXY(x, y)
	command = append(command, GetCleanLineFromCursor()...)
	i.m.Send(command)
}

func (i *Input) Repetition() {
	i.clearScreen()

	x, y := i.getAbsoluteXY(0)
	i.m.MoveCursorXY(x, y)
	i.m.Send(i.Value)

	if i.cursor {
		i.m.CursorOn()
	}
}

func (i *Input) Clear() {
	i.Value = []byte{}
	i.clearScreen()
}

func (i *Input) Activate() {
	x, y := i.getAbsoluteXY(0)
	i.m.MoveCursorXY(x, y)
	if i.cursor {
		i.m.CursorOn()
	}
}

func (i *Input) Deactivate() {
	i.m.CursorOff()
}
