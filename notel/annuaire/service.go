package annuaire

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func AnnuaireService(mntl *minigo.Minitel, userDB *databases.UsersDatabase) (op int) {
	_, op = NewPageList(mntl, userDB).Run()
	return
}
