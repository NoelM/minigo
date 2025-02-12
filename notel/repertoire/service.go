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
	m.SendVDT("static/annuaire.vdt")
}
