package annuaire

import (
	"fmt"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func NewPageDetail(mntl *minigo.Minitel, userDB *databases.UsersDatabase) *minigo.Page {
	detailPage := minigo.NewPage("notel:details", mntl, nil)

	var users []databases.User
	var userId int

	detailPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		var err error
		if users, err = userDB.LoadAnnuaireUsers(); err != nil {
			mntl.Print("Impossible de charger les utilisateurs")
			time.Sleep(2 * time.Second)
			return minigo.SommaireOp
		}

		mntl.MoveAt(2, 0)
		mntl.PrintAttributes("Annuaire", minigo.DoubleHauteur)

		mntl.Return(1)
		mntl.HLine(40, minigo.HCenter)

		displayUser(mntl, users[userId])
		displayHelpers(mntl, userId, len(users))

		return minigo.NoOp
	})

	detailPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	detailPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if userId == len(users)-1 {
			return nil, minigo.NoOp
		}

		mntl.CleanScreen()

		mntl.MoveAt(2, 0)
		mntl.PrintAttributes("Annuaire", minigo.DoubleHauteur)

		mntl.Return(1)
		mntl.HLine(40, minigo.HCenter)

		userId += 1
		displayUser(mntl, users[userId])
		displayHelpers(mntl, userId, len(users))

		return nil, minigo.NoOp
	})

	detailPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if userId == 0 {
			return nil, minigo.NoOp
		}

		mntl.CleanScreen()

		mntl.MoveAt(2, 0)
		mntl.PrintAttributes("Annuaire", minigo.DoubleHauteur)

		mntl.Return(1)
		mntl.HLine(40, minigo.HCenter)

		userId -= 1
		displayUser(mntl, users[userId])
		displayHelpers(mntl, userId, len(users))

		return nil, minigo.NoOp
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

func displayHelpers(mntl *minigo.Minitel, pos, maxPos int) {
	mntl.MoveAt(23, 0)

	if pos != 0 {
		mntl.Helper(fmt.Sprintf("Page %d/%d", pos, maxPos), "RETOUR", minigo.FondBleu, minigo.CaractereBlanc)
	}
	if pos+1 != maxPos {
		mntl.HelperRight(fmt.Sprintf("Page %d/%d", pos+2, maxPos), "SUITE", minigo.FondBleu, minigo.CaractereBlanc)
	}
}
