package infos

import "github.com/NoelM/minigo"

func ServiceInfo(m *minigo.Minitel) int {
HOME:
	serv, op := NewHomePage(m).Run()
	if op != minigo.EnvoiOp {
		return op
	}

	_, op = NewPageInfo(m, serv).Run()
	if op == minigo.SommaireOp {
		goto HOME
	}

	return op
}
