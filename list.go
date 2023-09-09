package minigo

import "fmt"

type List struct {
	mntl  *Minitel
	items []string
	refX  int
	refY  int
}

func NewList(mntl *Minitel, items []string) *List {
	return &List{
		mntl:  mntl,
		items: items,
		refX:  1,
		refY:  8,
	}
}

func (l *List) Display() {
	for i := 0; i < len(l.items); i += 1 {
		l.mntl.WriteAttributes(GrandeurNormale, InversionFond)
		l.mntl.WriteStringXY(l.refX, l.refY+2*i, fmt.Sprintf(" %d ", i+1))
		l.mntl.WriteAttributes(FondNormal)
		l.mntl.WriteStringXY(l.refX+4, l.refY+2*i, l.items[i])
	}
}
