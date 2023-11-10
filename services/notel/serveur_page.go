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
		mntl.WriteStringAt(2, 10, "NOTEL")
		mntl.WriteAttributes(minigo.GrandeurNormale, minigo.FondNormal)
		mntl.WriteStringLeft(3, "Mardi 7 Novembre 2023")

		messages := []string{
			"* Désormais une page de connexion sur NOTEL !",
			"* Arrivée des actualités comme nouveau service, basé sur les flux RSS de France Info",
			"* De nouveaux services à venir : un service de micro-blog, la gestion de son compte et les statistiques du serveur",
			"* Beaucoup de travail sur 'minigo' avec une page TODO des sujets à traiter",
		}

		line := 5
		for _, msg := range messages {
			for _, l := range minigo.WrapperLargeurNormale(msg) {
				mntl.WriteStringLeft(line, l)
				line += 1
			}
			line += 1
		}

		mntl.WriteHelperLeft(24, "Menu NOTEL", "SOMMAIRE")
		return minigo.NoOp
	})

	infoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	return infoPage
}
