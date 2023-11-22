package minigo

import "fmt"

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
	colAlign := 0
	for _, key := range l.orderedKeys {
		value := l.items[key]

		l.mntl.WriteAttributes(GrandeurNormale, InversionFond)
		l.mntl.WriteStringAt(line, colAlign+l.refCol, fmt.Sprintf(" %s ", key))

		l.mntl.WriteAttributes(FondNormal)
		l.mntl.WriteStringAt(line, colAlign+l.refCol+len(key)+3, value)

		line += l.brk
		if line >= l.maxRow {
			line = l.refRow
			colAlign = 20
		}
	}
}
