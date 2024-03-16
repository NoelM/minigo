package annuaire

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func AnnuaireService(mntl *minigo.Minitel, userDB *databases.UsersDatabase) int {
HOME:
	user, op := NewPageList(mntl, userDB).Run()
	if op != minigo.EnvoiOp {
		return minigo.SommaireOp
	}

	_, op = NewPageDetail(mntl, userDB, user["user"]).Run()
	if op == minigo.SommaireOp {
		goto HOME
	}

	return minigo.SommaireOp
}
