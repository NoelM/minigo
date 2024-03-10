package profil

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

const (
	mdpId = iota
	supprId
)

const (
	mdpKey   = "*MDP"
	supprKey = "*SUP"
)

var ServIdMap = map[string]int{
	mdpKey:   mdpId,
	supprKey: supprId,
}

func ProfilService(mntl *minigo.Minitel, userDB *databases.UsersDatabase, pseudo string) int {
HOME:
	op := RunPageProfil(mntl, userDB, pseudo)
	switch op {
	case minigo.SommaireOp:
		return op
	case mdpId:
		goto MDP
	case supprId:
		goto SUPPR
	default:
		goto HOME
	}

MDP:
	return RunMDPPage(mntl, userDB, pseudo)

SUPPR:
	return RunSupprPage(mntl, userDB, pseudo)
}
