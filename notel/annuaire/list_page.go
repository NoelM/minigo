package annuaire

import (
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func NewPageList(mntl *minigo.Minitel, userDB *databases.UsersDatabase) *minigo.Page {
	listPage := minigo.NewPage("notel:list", mntl, nil)

	var users []databases.User
	var userId int

	listPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		var err error
		if users, err = userDB.LoadAnnuaireUsers(); err != nil {
			mntl.Print("Impossible de charger les utilisateurs")
			time.Sleep(2 * time.Second)
			return minigo.SommaireOp
		}

		mntl.Attributes(minigo.CaractereBlanc, minigo.FondBleu)
		mntl.Repeat(0x20, 40)
		mntl.Attributes(minigo.CaractereBlanc, minigo.FondBleu)
		mntl.Repeat(0x20, 40)
		mntl.Attributes(minigo.CaractereBlanc, minigo.FondBleu)
		mntl.Repeat(0x20, 40)

		mntl.MoveAt(2, 0)
		mntl.Attributes(minigo.CaractereBlanc, minigo.FondBleu, minigo.DoubleHauteur)
		mntl.PrintCenter(" Annuaire ")

		mntl.Attributes(minigo.GrandeurNormale)

		userId = displayList(mntl, users, userId)

		return minigo.NoOp
	})

	listPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	listPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
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

	listPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
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

	return listPage
}

func displayList(m *minigo.Minitel, users []databases.User, userId int) int {

	m.MoveAt(4, 0)

	m.ModeG1()
	m.Attributes(minigo.CaractereVert)
	for i := 0; i < 21; i += 1 {
		m.Repeat(0x5f, 40)
	}

	m.MoveAt(5, 2)
	for i := userId; i < userId+6; i += 1 {
		if i >= len(users) {
			break
		}

		m.Attributes(minigo.FondBleu, minigo.CaractereBlanc, minigo.DoubleLargeur)
		m.Printf(" %d ", i+1)

		m.Attributes(minigo.FondVert, minigo.CaractereNoir)
		m.Printf(" %s", users[i].Nick)

		txt := minigo.WrapperGenerique(users[i].Bio, 26)
		m.ReturnCol(1, 9)
		m.Attributes(minigo.FondVert, minigo.GrandeurNormale)
		m.Printf(" %s...", txt[0])

		m.ReturnCol(2, 2)
	}

	return userId + 5
}
