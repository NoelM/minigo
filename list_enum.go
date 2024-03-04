package minigo

import "fmt"

func NewListEnum(mntl *Minitel, items []string, row, col, maxRow, brk int) *List {
	list := NewList(mntl, row, col, maxRow, brk)

	for i, val := range items {
		list.AppendItem(fmt.Sprintf("%d", i+1), val)
	}

	return list
}
