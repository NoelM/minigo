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
	return &List{
		mntl:        mntl,
		items:       make(map[string]string),
		orderedKeys: make([]string, 0),
		refRow:      row,
		refCol:      col,
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

		l.mntl.Attributes(FondBlanc, CaractereNoir)
		l.mntl.Print(" ")

		l.mntl.Print(key)

		l.mntl.Attributes(FondNormal, CaractereBlanc)
		l.mntl.Print(" ")

		l.mntl.Right(1)

		l.mntl.Attributes(FondNormal)
		l.mntl.Print(value)

		line += l.brk
		if line >= l.maxRow && colAlign == 0 {
			line = l.refRow
			colAlign = 20
			l.mntl.MoveAt(l.refRow, l.refCol+colAlign)
		} else {
			l.mntl.Return(l.brk)
			l.mntl.Right(l.refCol + colAlign)
		}
	}
}
