package main

import (
	"github.com/NoelM/minigo"
)

func NewServeurPage(mntl *minigo.Minitel) *minigo.Page {
	infoPage := minigo.NewPage("serveur", mntl, nil)

	infoPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.CursorOff()

		mntl.MoveAt(2, 1)
		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteString("Infos de")
		mntl.WriteAttributes(minigo.DoubleGrandeur, minigo.InversionFond)

		mntl.MoveRight(1)
		mntl.WriteString("NOTEL")
		mntl.WriteAttributes(minigo.GrandeurNormale, minigo.FondNormal)

		mntl.ReturnCol(2, 1)
		mntl.WriteString("Lundi 4 mars 2024")

		messages := []string{
			"* Refacto du code pour mieux utiliser le positionnement et les attribus de couleurs et fonds. Rendre compatible avec Minitel JS sur le web",
			"* Travail pour rendre les pages avec texte basées sur DB",
			"* Fichiers de configuration pour simplifier l'exécution du serveur",
		}

		line := 6
		for _, msg := range messages {
			for _, l := range minigo.WrapperLargeurNormale(msg) {
				mntl.WriteStringLeftAt(line, l)
				line += 1
			}
			line += 1
		}

		mntl.PrintHelperLeftAt(24, "Menu NOTEL", "SOMMAIRE")
		return minigo.NoOp
	})

	infoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	return infoPage
}
