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
		mntl.Attributes(minigo.DoubleHauteur)
		mntl.Print("Serveur")
		mntl.Attributes(minigo.DoubleGrandeur, minigo.InversionFond)

		mntl.Right(1)
		mntl.Print("NOTEL")
		mntl.Attributes(minigo.GrandeurNormale, minigo.FondNormal)

		mntl.Return(1)
		mntl.HLine(40, minigo.HCenter)

		mntl.HelperLeftAt(24, "Menu NOTEL", "SOMMAIRE")
		return minigo.NoOp
	})

	infoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	return infoPage
}
