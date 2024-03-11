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

		mntl.MoveAt(2, 0)
		mntl.Attributes(minigo.DoubleHauteur, minigo.CaractereJaune)
		mntl.PrintCenter("Serveur NOTEL")
		mntl.Attributes(minigo.GrandeurNormale)

		mntl.Return(1)
		mntl.HLine(40, minigo.HCenter)

		mntl.RectAt(4, 1, 38, 8)

		mntl.MoveAt(5, 0)
		mntl.Attributes(minigo.CaractereVert)
		mntl.PrintCenter("Stats de la semaine")

		mntl.ReturnCol(1, 2)
		mntl.HLine(36, minigo.Top)
		mntl.Attributes(minigo.CaractereBlanc)

		if r, err := Request(ConnectWeekly); err == nil {
			mntl.ReturnCol(1, 2)
			mntl.Print("Connexions")
			mntl.Right(2)

			f, _ := strconv.ParseFloat(r.Data.Result[0].Value[1].(string), 64)
			mntl.PrintAttributes(fmt.Sprintf("%.0f", f), minigo.CaractereVert)
		}

		if r, err := Request(DurationWeekly); err == nil {
			mntl.ReturnCol(1, 2)
			mntl.Print("Durée")
			mntl.Right(7)

			f, _ := strconv.ParseFloat(r.Data.Result[0].Value[1].(string), 64)
			mntl.PrintAttributes(fmt.Sprintf("%.0f H", f/3600.), minigo.CaractereVert)
		}

		if r, err := Request(MessagesWeekly); err == nil {
			mntl.ReturnCol(1, 2)
			mntl.Print("Messages")
			mntl.Right(4)

			f, _ := strconv.ParseFloat(r.Data.Result[0].Value[1].(string), 64)
			mntl.PrintAttributes(fmt.Sprintf("%.0f", f), minigo.CaractereVert)
		}

		mntl.RectAt(13, 1, 38, 7)

		mntl.MoveAt(14, 0)
		mntl.Attributes(minigo.CaractereMagenta)
		mntl.PrintCenter("Direct Live")

		mntl.ReturnCol(1, 2)
		mntl.HLine(36, minigo.Top)

		mntl.Attributes(minigo.CaractereBlanc)

		if r, err := Request(CPULoad); err == nil {
			mntl.ReturnCol(1, 2)
			mntl.Print("Processeur")
			mntl.Right(2)

			f, _ := strconv.ParseFloat(r.Data.Result[0].Value[1].(string), 64)
			mntl.PrintAttributes(fmt.Sprintf("%.1f %%", f), minigo.CaractereMagenta)
		}

		if r, err := Request(CPUTemp); err == nil {
			mntl.ReturnCol(1, 2)
			mntl.Print("Température")
			mntl.Right(1)

			f, _ := strconv.ParseFloat(r.Data.Result[0].Value[1].(string), 64)
			mntl.PrintAttributes(fmt.Sprintf("%.1f °C", f), minigo.CaractereMagenta)
		}

		mntl.HelperLeftAt(24, "Menu NOTEL", "SOMMAIRE")
		return minigo.NoOp
	})

	infoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	return infoPage
}
