package superchat

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/metrics"
)

func HelpPage(minitel *minigo.Minitel, mtr *metrics.Metrics) *minigo.Page {
	helpPage := minigo.NewPage("superchat: help", minitel, nil)

	helpPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.CursorOff()
		mntl.ModeG0()

		mntl.MoveAt(2, 0)
		mntl.Attributes(minigo.DoubleGrandeur)
		mntl.PrintCenter("^..^ SuperChat ^..^")
		mntl.Attributes(minigo.GrandeurNormale)
		mntl.Return(2)

		mntl.HLine(40, minigo.HCenter)
		mntl.Return(1)

		mntl.Print("SuperChat est en bêta: il y a des bugs.")
		mntl.Return(2)

		mntl.Print("Mode ")
		mntl.Attributes(minigo.FondVert, minigo.CaractereNoir)
		mntl.Print(" EDITION ")
		mntl.Attributes(minigo.FondNoir, minigo.CaractereBlanc)
		mntl.Print(" actif par défaut")
		mntl.Return(1)
		mntl.Print("Activez avec le bouton REPETITION")
		mntl.Return(1)

		mntl.Print("- Pour écrire des messages")
		mntl.Return(1)
		mntl.Print("- Charger les messages en direct")
		mntl.Return(2)

		mntl.Print("Mode ")
		mntl.Attributes(minigo.FondVert, minigo.CaractereNoir)
		mntl.Print(" NAVIGATION")
		mntl.Attributes(minigo.FondNormal, minigo.CaractereBlanc)
		mntl.Print(" ")
		mntl.Return(1)
		mntl.Print("Pour naviguer dans les messages avec")
		mntl.Return(1)

		mntl.Print("- SUITE, message suivant")
		mntl.Return(1)
		mntl.Print("- RETOUR, message précédent")
		mntl.Return(2)

		mntl.HLine(40, minigo.HCenter)
		mntl.Return(1)

		logged := mtr.ListLogged()

		if len(logged) == 1 {
			mntl.Print("Connecté:")
		} else {
			mntl.Print("Connectés:")
		}

		for _, loggedNick := range logged {
			mntl.Attributes(minigo.InversionFond)
			mntl.Printf(" %s ", loggedNick)
			mntl.Attributes(minigo.FondNormal)
			mntl.Right(1)
		}

		mntl.HelperRightAt(24, "Aller au Chat", "SOMMAIRE")

		return minigo.NoOp
	})

	helpPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	return helpPage
}
