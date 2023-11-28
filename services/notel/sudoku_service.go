package main

import "github.com/NoelM/minigo"

func SudokuService(mntl *minigo.Minitel) int {
LEVEL:
	level, op := NewPageLevel(mntl).Run()
	if op != minigo.EnvoiOp {
		return minigo.SommaireOp
	}

	_, op = NewPageGame(mntl, level).Run()
	if op == minigo.EnvoiOp {
		goto LEVEL
	}

	return op
}
