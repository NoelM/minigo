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

func (i *Input) AppendKey(key byte) {
	totalLen := len(i.pre) + 1 + len(i.Value)
	locY := totalLen / i.width
	locX := totalLen % i.width

	i.Value = append(i.Value, key)

	command := GetMoveCursorXY(i.refX+locX+1, i.refY+locY+totalLen)
	command = append(command, key)
	i.m.Send(command)
}

func (i *Input) Correction() {
	if len(i.Value) == 0 {
		return
	}

	totalLen := len(i.pre) + 1 + len(i.Value)
	locY := (totalLen - 1) / i.width
	locX := (totalLen - 1) % i.width

	i.Value = i.Value[:len(i.Value)-1]

	command := GetMoveCursorXY(i.refX+locX+1, i.refY+locY)
	command = append(command, GetCleanLineFromCursor()...)
	i.m.Send(command)
}

func (i *Input) clearScreen() {
	command := []byte{}

	for row := 0; row < i.height; row += 1 {
		command = GetMoveCursorXY(i.refX, i.refY)

		// TODO: handle input with a width < rowWidth
		command = append(command, GetCleanScreenFromCursor()...)
	}
	command = append(command, EncodeMessage(i.pre)...)
	command = append(command, GetMoveCursorRight(1)...)
	command = append(command, []byte(i.pre)...)
	command = append(command, GetMoveCursorRight(1)...)
	i.m.Send(command)
}

func (i *Input) Print() {
	i.clearScreen()
	i.m.Send(i.Value)
}

func (i *Input) Clear() {
	i.Value = []byte{}
	i.clearScreen()
}
