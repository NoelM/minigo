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
	m.SendVDT("static/annuaire.vdt")
}
