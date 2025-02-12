package repertoire

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func RepertoireService(m *minigo.Minitel, userDB *databases.UsersDatabase) int {
HOME:
	user, op := NewPageList(m, userDB).Run()
	if op != minigo.EnvoiOp {
		return op
	}

	_, op = NewPageDetail(m, userDB, user["user"]).Run()
	if op == minigo.SommaireOp {
		goto HOME
	}

	return op
}

func printRepertoireHeader(m *minigo.Minitel) {
	m.MoveAt(1, 0)
	m.Attributes(minigo.FondCyan, minigo.CaractereNoir)
	m.Repeat(' ', 40)
	m.Return(1)
	m.Attributes(minigo.DoubleGrandeur, minigo.FondCyan, minigo.CaractereNoir)
	m.Print("  RÉPERTOIRE")
	m.Repeat(' ', 28)
	m.Attributes(minigo.GrandeurNormale, minigo.FondCyan, minigo.CaractereNoir)
	m.ReturnUp(2)
	m.Repeat(' ', 40)
	m.Attributes(minigo.FondNormal, minigo.CaractereBlanc)
	m.Print(" ")
}
