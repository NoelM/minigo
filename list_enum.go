package minigo

import "fmt"

type ListEnum struct {
	mntl        *Minitel
	items       []string
	refX        int
	refY        int
	entryHeight int
}

func NewListEnum(mntl *Minitel, items []string) *ListEnum {
	return &ListEnum{
		mntl:        mntl,
		items:       items,
		refX:        1,
		refY:        8,
		entryHeight: 2,
	}
}

func (l *ListEnum) SetXY(x, y int) {
	l.refX = x
	l.refY = y
}

func (l *ListEnum) AppendItem(item string) {
	l.items = append(l.items, item)
}

func (l *ListEnum) SetEntryHeight(h int) {
	l.entryHeight = h
}

func (l *ListEnum) Display() {
	for i := 0; i < len(l.items); i += 1 {
		l.mntl.WriteAttributes(GrandeurNormale, InversionFond)
		l.mntl.WriteStringAt(l.refY+l.entryHeight*i, l.refX, fmt.Sprintf(" %d ", i+1))
		l.mntl.WriteAttributes(FondNormal)
		l.mntl.WriteStringAt(l.refY+l.entryHeight*i, l.refX+4, l.items[i])
	}
}
