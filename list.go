package minigo

type List struct {
	mntl        *Minitel
	orderedKeys []string
	items       map[string]string
	refCol      int
	refRow      int
	maxRow      int
	brk         int
}

func NewList(mntl *Minitel, row, col, maxRow, brk int) *List {
	// Because the background iversion require a blank space
	// to be activated, we substract 1 to the refCol
	// in order to print a blank SPACE
	rCol := col - 1
	if rCol < 0 {
		rCol = 0
	}

	return &List{
		mntl:        mntl,
		items:       make(map[string]string),
		orderedKeys: make([]string, 0),
		refRow:      row,
		refCol:      rCol,
		maxRow:      maxRow,
		brk:         brk,
	}
}

func (l *List) AppendItem(key, value string) {
	l.orderedKeys = append(l.orderedKeys, key)
	l.items[key] = value
}

func (l *List) Display() {
	line := l.refRow
	l.mntl.MoveAt(l.refRow, l.refCol)

	colAlign := 0
	for _, key := range l.orderedKeys {
		value := l.items[key]

		l.mntl.WriteAttributes(FondBlanc, CaractereNoir)
		l.mntl.WriteString("  ")
		//                  |^ white
		//                  |- blank

		l.mntl.WriteString(key)

		l.mntl.WriteAttributes(FondNormal, CaractereBlanc)
		l.mntl.WriteString(" ")
		//                  ^ white

		l.mntl.MoveRight(3)

		l.mntl.WriteAttributes(FondNormal)
		l.mntl.WriteString(value)

		line += l.brk
		if line >= l.maxRow {
			line = l.refRow
			colAlign = 20
		}

		l.mntl.Return(l.brk)
		l.mntl.MoveLeft(l.refCol + colAlign)
	}
}
