package main

import (
	"fmt"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func NewPageAnnuaire(mntl *minigo.Minitel, userDB *databases.UsersDatabase) *minigo.Page {
	infoPage := minigo.NewPage("notel:annuaire", mntl, nil)

	var users []databases.User
	var userId int

	infoPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		var err error
		if users, err = userDB.LoadAnnuaireUsers(); err != nil {
			mntl.WriteString("Impossible de charger les utilisateurs")
			time.Sleep(2 * time.Second)
			return minigo.SommaireOp
		}

		mntl.MoveAt(2, 0)
		mntl.WriteStringWithAttributes("Annuaire", minigo.DoubleHauteur)

		mntl.Return(1)
		mntl.HLine(40, minigo.HCenter)

		displayUser(mntl, users[userId])
		displayHelpers(mntl, userId, len(users))

		return minigo.NoOp
	})

	infoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	infoPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if userId == len(users)-1 {
			return nil, minigo.NoOp
		}

		mntl.CleanScreen()

		mntl.MoveAt(2, 0)
		mntl.WriteStringWithAttributes("Annuaire", minigo.DoubleHauteur)

		mntl.Return(1)
		mntl.HLine(40, minigo.HCenter)

		userId += 1
		displayUser(mntl, users[userId])
		displayHelpers(mntl, userId, len(users))

		return nil, minigo.NoOp
	})

	infoPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if userId == 0 {
			return nil, minigo.NoOp
		}

		mntl.CleanScreen()

		mntl.MoveAt(2, 0)
		mntl.WriteStringWithAttributes("Annuaire", minigo.DoubleHauteur)

		mntl.Return(1)
		mntl.HLine(40, minigo.HCenter)

		userId -= 1
		displayUser(mntl, users[userId])
		displayHelpers(mntl, userId, len(users))

		return nil, minigo.NoOp
	})

	return infoPage
}

func displayUser(mntl *minigo.Minitel, user databases.User) {
	mntl.MoveAt(5, 0)

	mntl.WriteStringWithAttributes(user.Nick, minigo.InversionFond, minigo.DoubleGrandeur)

	mntl.Return(1)
	mntl.HLine(40, minigo.HCenter)

	mntl.Return(1) // Row 6
	mntl.WriteAttributes(minigo.FondRouge)
	mntl.WriteRepeat(' ', 40)
	mntl.MoveUp(1)
	mntl.WriteStringCenter("=== Bio ===")

	mntl.Return(1)
	mntl.WriteAttributes(minigo.FondNoir)
	mntl.WriteString(" ")
	mntl.MoveLeft(1)

	for _, line := range minigo.WrapperLargeurNormale(user.Bio) {
		mntl.Return(1)
		mntl.WriteString(line)
	}

	mntl.Return(2)
	mntl.WriteAttributes(minigo.FondBleu)
	mntl.WriteRepeat(' ', 40)
	mntl.MoveUp(1)
	mntl.WriteStringCenter("=== Serveur Minitel ===")

	mntl.Return(2)
	mntl.WriteAttributes(minigo.FondNoir)
	mntl.WriteString(" ")
	mntl.MoveLeft(1)

	mntl.WriteString(user.Tel)

	mntl.Return(2) // Row 13
	mntl.WriteAttributes(minigo.FondJaune, minigo.CaractereNoir)
	mntl.WriteRepeat(' ', 40)
	mntl.MoveUp(1)
	mntl.WriteStringCenter("=== Lieu ===")

	mntl.Return(2)
	mntl.WriteAttributes(minigo.FondNoir, minigo.CaractereBlanc)
	mntl.WriteString(" ")
	mntl.MoveLeft(1)

	mntl.WriteString(user.Location)
}

func displayHelpers(mntl *minigo.Minitel, pos, maxPos int) {
	mntl.MoveAt(23, 0)

	if pos != 0 {
		mntl.PrintHelper(fmt.Sprintf("Page %d/%d", pos, maxPos), "RETOUR", minigo.FondBleu, minigo.CaractereBlanc)
	}
	if pos+1 != maxPos {
		mntl.PrintHelperRight(fmt.Sprintf("Page %d/%d", pos+2, maxPos), "SUITE", minigo.FondBleu, minigo.CaractereBlanc)
	}
}
