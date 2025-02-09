package superchat

import "github.com/NoelM/minigo"

func HelpPage(minitel *minigo.Minitel) *minigo.Page {
	helpPage := minigo.NewPage("superchat: welcome", minitel, nil)

	helpPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.CursorOff()
		mntl.ModeG0()

		mntl.MoveAt(2, 0)
		mntl.Attributes(minigo.DoubleGrandeur)
		mntl.PrintCenter("^..^ SuperChat ^..^")
		mntl.Attributes(minigo.GrandeurNormale)
		mntl.Return(1)

		mntl.PrintCenter("Page d'aide")
		mntl.Return(2)

		mntl.HLine(40, minigo.HCenter)
		mntl.Return(1)

		mntl.Print("SuperChat est en bêta: il y a des bugs.")
		mntl.Return(2)

		mntl.Print("Vous pouvez naviguer dans l'historique:")
		mntl.Return(1)

		mntl.Print("- SUITE : message suivant")
		mntl.Return(1)
		mntl.Print("- RETOUR : message précédent")
		mntl.Return(2)

		mntl.Print("Appuyez sur REPETITION :")
		mntl.Return(1)

		mntl.Print("- Pour revenir en mode édition")
		mntl.Return(1)
		mntl.Print("- Pour charger le dernier msg")
		mntl.Return(2)

		mntl.HLine(40, minigo.HCenter)
		mntl.Return(1)

		mntl.Print("Appuyez sur GUIDE n'importe quand pour revoir cette page.")

		mntl.HelperRightAt(24, "Aller au Chat", "SOMMAIRE")

		return minigo.NoOp
	})

	helpPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	return helpPage
}
