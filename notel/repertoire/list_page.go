package repertoire

import (
	"strconv"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

const usersPerPage = 6

func NewPageList(mntl *minigo.Minitel, userDB *databases.UsersDatabase) *minigo.Page {
	listPage := minigo.NewPage("notel:list", mntl, nil)

	var users []databases.User

	var pageId int
	var selectedUserId int64 = -1

	listPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.Reset()

		var err error
		if users, err = userDB.LoadAllUsers(); err != nil {
			mntl.Print("Impossible de charger les utilisateurs")
			time.Sleep(2 * time.Second)
			return minigo.SommaireOp
		}

		displayPage(mntl, users, usersPerPage, pageId)

		return minigo.NoOp
	})

	listPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	listPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key rune) {
		var err error
		selectedUserId, err = strconv.ParseInt(string(key), 10, 64)
		if err != nil {
			selectedUserId = -1
			return
		}
	})

	listPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if pageId == len(users)/usersPerPage {
			return nil, minigo.NoOp
		}

		pageId += 1
		displayPage(mntl, users, usersPerPage, pageId)

		return nil, minigo.NoOp
	})

	listPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if pageId == 0 {
			return nil, minigo.NoOp
		}

		pageId -= 1
		displayPage(mntl, users, usersPerPage, pageId)

		return nil, minigo.NoOp
	})

	listPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if selectedUserId > 0 && selectedUserId < 7 {
			return map[string]string{"user": users[selectedUserId-1].Nick}, minigo.EnvoiOp
		}
		return nil, minigo.NoOp
	})

	return listPage
}

func displayPage(m *minigo.Minitel, users []databases.User, usersPerPage, pageId int) {
	printRepertoireHeader(m)

	m.ModeG0()

	m.MoveAt(3, 35)
	m.Attributes(minigo.CaractereNoir)
	m.Printf("%d/%d ", pageId+1, len(users)/usersPerPage+1)

	displayList(m, users, pageId*usersPerPage)

	m.MoveAt(24, 0)
	m.Attributes(minigo.CaractereCyan)
	m.HelperRight("Numéro du profil + ", "ENVOI", minigo.FondCyan, minigo.CaractereNoir)
}

func displayList(m *minigo.Minitel, users []databases.User, userId int) {

	m.MoveAt(6, 2)
	for i := userId; i < userId+usersPerPage; i += 1 {
		if i >= len(users) {
			break
		}

		m.Attributes(minigo.CaractereCyan, minigo.DoubleLargeur, minigo.InversionFond)
		m.Printf(" %d ", i+1)

		m.Right(2)
		m.Attributes(minigo.FondNormal)
		m.Print(users[i].Nick)

		txt := minigo.Wrapper(users[i].Bio, 26)
		m.ReturnCol(1, 10)
		m.Attributes(minigo.CaractereCyan, minigo.GrandeurNormale)
		m.Printf("%s...", txt[0])

		m.ReturnCol(2, 2)
	}
}
