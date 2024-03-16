package annuaire

import (
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func NewPageDetail(mntl *minigo.Minitel, userDB *databases.UsersDatabase, nick string) *minigo.Page {
	detailPage := minigo.NewPage("notel:details", mntl, nil)

	detailPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		user, err := userDB.LoadUser(nick)
		if err != nil {
			mntl.Print("Impossible de charger l'utilisateur")
			time.Sleep(2 * time.Second)
			return minigo.SommaireOp
		}

		mntl.MoveAt(2, 0)
		mntl.PrintAttributes("Annuaire", minigo.DoubleHauteur)

		mntl.Return(1)
		mntl.HLine(40, minigo.HCenter)

		displayUser(mntl, user)
		displayHelpers(mntl)

		return minigo.NoOp
	})

	detailPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	return detailPage
}

func displayUser(mntl *minigo.Minitel, user databases.User) {
	mntl.MoveAt(5, 0)
	mntl.Attributes(minigo.FondBleu)
	mntl.Print(" PSEUDO")
	mntl.SendCAN()

	mntl.Return(2)
	mntl.Attributes(minigo.FondNoir)
	mntl.Print(" ")
	mntl.Left(1)

	mntl.PrintAttributes(user.Nick, minigo.DoubleLargeur)

	mntl.Return(2) // Row 6
	mntl.Attributes(minigo.FondBleu)
	mntl.Print(" BIO")
	mntl.SendCAN()

	mntl.Return(1)
	mntl.Attributes(minigo.FondNoir)
	mntl.Print(" ")
	mntl.Left(1)

	for _, line := range minigo.WrapperLargeurNormale(user.Bio) {
		mntl.Return(1)
		mntl.Print(line)
	}

	mntl.Return(2)
	mntl.Attributes(minigo.FondBleu)
	mntl.Print(" TÉLÉPHONE")
	mntl.SendCAN()

	mntl.Return(2)
	mntl.Attributes(minigo.FondNoir)
	mntl.Print(" ")
	mntl.Left(1)

	mntl.Print(user.Tel)

	mntl.Return(2) // Row 13
	mntl.Attributes(minigo.FondBleu)
	mntl.Print(" LIEU")
	mntl.SendCAN()

	mntl.Return(2)
	mntl.Attributes(minigo.FondNoir)
	mntl.Print(" ")
	mntl.Left(1)

	mntl.Print(user.Location)
}

func displayHelpers(mntl *minigo.Minitel) {
	mntl.MoveAt(24, 0)
	mntl.Helper("Liste des utilisateurs", "SOMMAIRE", minigo.FondBleu, minigo.CaractereBlanc)
}
