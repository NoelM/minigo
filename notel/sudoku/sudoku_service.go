package sudoku

import "github.com/NoelM/minigo"

func SudokuService(mntl *minigo.Minitel, login string) int {
LEVEL:
	level, op := RunPageLevel(mntl)
	if op != minigo.EnvoiOp {
		return minigo.SommaireOp
	}

	op = RunPageGame(mntl, login, level)
	if op == minigo.EnvoiOp {
		goto LEVEL
	}

	return op
}
