package minigo

import "fmt"

type List struct {
	mntl        *Minitel
	items       []string
	refX        int
	refY        int
	entryHeight int
}

func NewList(mntl *Minitel, items []string) *List {
	return &List{
		mntl:        mntl,
		items:       items,
		refX:        1,
		refY:        8,
		entryHeight: 2,
	}
}

func (l *List) SetXY(x, y int) {
	l.refX = x
	l.refY = y
}

func (l *List) AppendItem(item string) {
	l.items = append(l.items, item)
}

func (l *List) SetEntryHeight(h int) {
	l.entryHeight = h
}

func (l *List) Display() {
	for i := 0; i < len(l.items); i += 1 {
		l.mntl.WriteAttributes(GrandeurNormale, InversionFond)
		l.mntl.WriteStringXY(l.refX, l.refY+l.entryHeight*i, fmt.Sprintf(" %d ", i+1))
		l.mntl.WriteAttributes(FondNormal)
		l.mntl.WriteStringXY(l.refX+4, l.refY+l.entryHeight*i, l.items[i])
	}
}
