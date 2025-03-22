package annuaire

import "github.com/NoelM/minigo"

func AnnuaireService(mntl *minigo.Minitel, annuaireDbPath string) int {
	_, op := NewPageAnnuaire(mntl, annuaireDbPath).Run()

	return op
}
