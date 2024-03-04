package main

import (
	"github.com/NoelM/minigo"
)

func NewServeurPage(mntl *minigo.Minitel) *minigo.Page {
	infoPage := minigo.NewPage("serveur", mntl, nil)

	infoPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.CursorOff()

		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringLeftAt(2, "Infos de")
		mntl.WriteAttributes(minigo.DoubleGrandeur, minigo.InversionFond)
		mntl.WriteStringAt(2, 10, "NOTEL")
		mntl.WriteAttributes(minigo.GrandeurNormale, minigo.FondNormal)
		mntl.WriteStringLeftAt(4, "Vendredi 22 Décembre 2023")

		messages := []string{
			"- Beaucoup de nouveautés sur le serveur : un sudoku, un nouvel affichage des messages sur le chat, et beaucoup de travail caché dans le code",
			"- Encore quelques bugs connus : avec la PCE la gestion des déconnexions est plus complexe, parfois le compteur de connectés est faux :(",
			"- Toujours en plans : le service de micro blog, c'est pas simple à faire ! Et la navigation dans les message du chat avec Retour/Suite",
		}

		line := 6
		for _, msg := range messages {
			for _, l := range minigo.WrapperLargeurNormale(msg) {
				mntl.WriteStringLeftAt(line, l)
				line += 1
			}
			line += 1
		}

		mntl.WriteHelperLeftAt(24, "Menu NOTEL", "SOMMAIRE")
		return minigo.NoOp
	})

	infoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	return infoPage
}
