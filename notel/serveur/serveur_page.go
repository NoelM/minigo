package serveur

import (
	"fmt"
	"strconv"

	"github.com/NoelM/minigo"
)

func NewServeurPage(mntl *minigo.Minitel) *minigo.Page {
	infoPage := minigo.NewPage("serveur", mntl, nil)

	infoPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.CursorOff()

		mntl.MoveAt(2, 1)
		mntl.Attributes(minigo.DoubleHauteur)
		mntl.Print("Serveur")
		mntl.Attributes(minigo.DoubleGrandeur, minigo.InversionFond)

		mntl.Right(1)
		mntl.Print("NOTEL")
		mntl.Attributes(minigo.GrandeurNormale, minigo.FondNormal)

		mntl.Return(1)
		mntl.HLine(40, minigo.HCenter)

		mntl.Return(2)
		mntl.PrintAttributes("→ Cette semaine", minigo.DoubleHauteur)
		mntl.Return(2)

		if r, err := Request(ConnectWeekly); err == nil {
			mntl.PrintAttributes(" Connexions ", minigo.InversionFond)

			f, _ := strconv.ParseFloat(r.Data.Result[0].Value[1].(string), 64)
			mntl.PrintAttributes(fmt.Sprintf(" %.0f", f), minigo.CaractereVert)

			mntl.Return(1)
			mntl.PrintAttributes("(connexions de tests comprises)", minigo.CaractereRouge)
		}

		if r, err := Request(DurationWeekly); err == nil {
			mntl.Return(2)
			mntl.PrintAttributes(" Durées ", minigo.InversionFond)

			f, _ := strconv.ParseFloat(r.Data.Result[0].Value[1].(string), 64)
			mntl.PrintAttributes(fmt.Sprintf(" %.1f heures", f/3600.), minigo.CaractereVert)
		}

		if r, err := Request(MessagesWeekly); err == nil {
			mntl.Return(2)
			mntl.PrintAttributes(" Messages ", minigo.InversionFond)

			f, _ := strconv.ParseFloat(r.Data.Result[0].Value[1].(string), 64)
			mntl.PrintAttributes(fmt.Sprintf(" %.0f", f), minigo.CaractereVert)
		}

		mntl.Return(2)
		mntl.HLine(40, minigo.HCenter)

		mntl.Return(2)
		mntl.PrintAttributes("→ En ce moment", minigo.DoubleHauteur)

		if r, err := Request(CPULoad); err == nil {
			mntl.Return(2)
			mntl.PrintAttributes(" Processeur ", minigo.InversionFond)

			f, _ := strconv.ParseFloat(r.Data.Result[0].Value[1].(string), 64)
			mntl.PrintAttributes(fmt.Sprintf(" %.1f %%", f), minigo.CaractereVert)
		}

		if r, err := Request(CPUTemp); err == nil {
			mntl.Return(2)
			mntl.PrintAttributes(" Température ", minigo.InversionFond)

			f, _ := strconv.ParseFloat(r.Data.Result[0].Value[1].(string), 64)
			mntl.PrintAttributes(fmt.Sprintf(" %.1f °C", f), minigo.CaractereVert)
		}

		mntl.HelperLeftAt(24, "Menu NOTEL", "SOMMAIRE")
		return minigo.NoOp
	})

	infoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	return infoPage
}
