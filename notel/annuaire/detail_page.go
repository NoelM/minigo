package annuaire

import (
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func NewPageDetail(mntl *minigo.Minitel, userDB *databases.UsersDatabase, nick string) *minigo.Page {
	detailPage := minigo.NewPage("notel:details", mntl, nil)

	detailPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.Reset()

		user, err := userDB.LoadUser(nick)
		if err != nil {
			mntl.Print("Impossible de charger l'utilisateur")
			time.Sleep(2 * time.Second)
			return minigo.SommaireOp
		}

		printAnnuaireHeader(mntl)
		printUserDetails(mntl, user)
		printHelpers(mntl)

		return minigo.NoOp
	})

	detailPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	return detailPage
}

func printUserDetails(mntl *minigo.Minitel, user databases.User) {
	mntl.MoveAt(6, 0)

	mntl.Attributes(minigo.CaractereVert, minigo.DoubleLargeur)
	mntl.PrintCenter(user.Nick)

	mntl.Return(2)
	mntl.Attributes(minigo.FondVert, minigo.CaractereNoir, minigo.GrandeurNormale)
	mntl.Print(" → BIO")
	mntl.SendCAN()

	mntl.Attributes(minigo.CaractereVert)
	for _, line := range minigo.WrapperGenerique(user.Bio, 37) {
		mntl.ReturnCol(1, 1)
		mntl.Print(line)
	}

	mntl.Return(2)
	mntl.Attributes(minigo.FondVert, minigo.CaractereNoir)
	mntl.Print(" → SERVEUR MINITEL")
	mntl.SendCAN()

	mntl.ReturnCol(1, 1)
	mntl.Attributes(minigo.CaractereVert)

	mntl.Print(user.Tel)

	mntl.Return(2) // Row 13
	mntl.Attributes(minigo.FondVert, minigo.CaractereNoir)
	mntl.Print(" → LIEU")
	mntl.SendCAN()

	mntl.ReturnCol(1, 1)
	mntl.Attributes(minigo.CaractereVert)

	mntl.Print(user.Location)
}

func printHelpers(mntl *minigo.Minitel) {
	mntl.MoveAt(24, 0)
	mntl.HelperRight("Liste des utilisateurs", "SOMMAIRE", minigo.FondVert, minigo.CaractereNoir)
}
