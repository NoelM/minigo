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
		mntl.WriteStringLeft(3, "Lundi 27 Novembre 2023")

		messages := []string{
			"* Ca y est ! La PCE est complète, si votre Minitel reporte plus de 5 erreurs de parités par minute, la PCE s'active automatiquement",
			"* Bug connu: pour l'instant il faut se déconnecter avec deux appuis sur Connexion/Fin",
			"* Le reste: pour la PCE j'avais mis en pause les statistiques d'usage du serveur, la modification de se compte et un service de micro-blog",
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