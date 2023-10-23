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
		mntl.WriteStringLeft(2, "Infos de")
		mntl.WriteAttributes(minigo.DoubleGrandeur, minigo.InversionFond)
		mntl.WriteStringAt(10, 2, "NOTEL")
		mntl.WriteAttributes(minigo.GrandeurNormale, minigo.FondNormal)
		mntl.WriteStringLeft(3, "Lundi 23 Octobre 2023")

		mntl.WriteStringLeft(5, "- Arrivée du multivoies le 28/10")
		mntl.WriteStringLeft(6, "  2 modems US Robotics arrivés")
		mntl.WriteStringLeft(7, "- Lancement samedi de statistiques")
		mntl.WriteStringLeft(8, "  de connexion sur le serveur")
		mntl.WriteStringLeft(9, "- Service d'actualités basées")
		mntl.WriteStringLeft(10, "  sur des flux RSS.")

		mntl.WriteHelperLeft(24, "Menu NOTEL", "SOMMAIRE")
		return minigo.NoOp
	})

	infoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	return infoPage
}
