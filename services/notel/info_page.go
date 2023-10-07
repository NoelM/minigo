package main

import (
	"github.com/NoelM/minigo"
)

func NewInfoPage(mntl *minigo.Minitel) *minigo.Page {
	infoPage := minigo.NewPage("info", mntl, nil)

	infoPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.CursorOff()

		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringLeft(2, "Infos de")
		mntl.WriteAttributes(minigo.DoubleGrandeur, minigo.InversionFond)
		mntl.WriteStringAt(10, 2, "NOTEL")
		mntl.WriteAttributes(minigo.GrandeurNormale, minigo.FondNormal)
		mntl.WriteStringLeft(3, "Samedi 7 Octobre 2023")

		mntl.WriteStringLeft(5, "- PAPT, modems, et switch bien arrivés,")
		mntl.WriteStringLeft(6, "  plus qu'à configurer la bête en")
		mntl.WriteStringLeft(7, "  multi-voies")
		mntl.WriteStringLeft(8, "- Nouvelle météo, plus détaillée et")
		mntl.WriteStringLeft(9, "  plus simple à lire")
		mntl.WriteStringLeft(10, "- DB de Chat et nouvel affichage")
		mntl.WriteStringLeft(11, "  nouvel affichage des dates")
		mntl.WriteStringLeft(12, "- Beaucoup d'améliorations de la")
		mntl.WriteStringLeft(13, "  librairie 'minigo'")
		mntl.WriteStringLeft(14, "- Prochaine étape, unifier la navi-")
		mntl.WriteStringLeft(15, "  gation entre les pages")

		mntl.WriteHelperLeft(24, "Menu NOTEL", "SOMMAIRE")
		return minigo.NoOp
	})

	infoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	return infoPage
}
