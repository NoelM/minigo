package annuaire

import (
	"encoding/json"
	"os"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/logs"
)

type AnnuaireEntry struct {
	Name  string `json:"name"`
	Hours string `json:"hours,omitempty"`
	Phone string `json:"phone"`
}

func NewPageAnnuaire(mntl *minigo.Minitel, annuaireDbPath string) *minigo.Page {
	page := minigo.NewPage("annuaire", mntl, nil)

	page.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.Reset()
		mntl.CursorOff()

		servers := []AnnuaireEntry{}
		if data, err := os.ReadFile(annuaireDbPath); err != nil {
			logs.ErrorLog("unable to open Annuaire DB: %s\n", err)
			return minigo.SommaireOp

		} else {
			if err := json.Unmarshal(data, &servers); err != nil {
				logs.ErrorLog("unable to unmarshal Annuaire DB: %s\n", err)
				return minigo.SommaireOp
			}
		}

		printRepertoireHeader(mntl)
		mntl.ModeG0()

		mntl.MoveAt(5, 0)

		isGreen := true
		for _, item := range servers {
			mntl.Return(1)
			mntl.Right(1)

			if isGreen {
				mntl.Attributes(minigo.FondVert, minigo.CaractereNoir)
			} else {
				mntl.Attributes(minigo.FondNormal, minigo.CaractereVert)
			}
			isGreen = !isGreen

			mntl.Printf(" %s", item.Name)

			// X_[NAME]_..._[01 02 03 04 05]_X
			// 2 |          |15
			//   len(name)

			mntl.Repeat(' ', 21-len(item.Name))
			mntl.Printf("%s ", item.Phone)
		}

		return minigo.NoOp
	})

	page.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	return page
}

func printRepertoireHeader(m *minigo.Minitel) {
	m.SendVDT("static/annuaire.vdt")
}
