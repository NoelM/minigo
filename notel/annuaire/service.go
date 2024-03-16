package annuaire

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func AnnuaireService(m *minigo.Minitel, userDB *databases.UsersDatabase) int {
HOME:
	user, op := NewPageList(m, userDB).Run()
	if op != minigo.EnvoiOp {
		return minigo.SommaireOp
	}

	_, op = NewPageDetail(m, userDB, user["user"]).Run()
	if op == minigo.SommaireOp {
		goto HOME
	}

	return minigo.SommaireOp
}

func printAnnuaireHeader(m *minigo.Minitel) {
	m.CleanScreen()
	m.MoveAt(1, 0)

	m.Attributes(minigo.CaractereBlanc, minigo.FondBleu)
	m.Print(" ")
	m.HLine(38, minigo.Top)
	m.SendCAN()
	m.Return(1)

	m.Attributes(minigo.CaractereBlanc, minigo.FondBleu)
	m.Print(" ")
	m.HLine(38, minigo.HCenter)
	m.SendCAN()
	m.Return(1)

	m.Attributes(minigo.CaractereBlanc, minigo.FondBleu)
	m.Print(" ")
	m.HLine(38, minigo.Bottom)
	m.SendCAN()

	m.MoveAt(2, 0)
	m.Attributes(minigo.CaractereBlanc, minigo.FondBleu, minigo.DoubleHauteur)
	m.PrintCenter(" Annuaire ")

	m.Attributes(minigo.GrandeurNormale)
}
