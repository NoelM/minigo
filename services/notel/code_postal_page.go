package main

import (
	"github.com/NoelM/minigo"
)

func NewCodePostalPage(mntl *minigo.Minitel) *minigo.Page {
	codePostalPage := minigo.NewPage("code_postal", mntl, nil)

	codePostalPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringXY(12, 2, "Prévisions Météo")
		mntl.WriteAttributes(minigo.GrandeurNormale)

		mntl.WriteStringXY(1, 4, "CODE POSTAL: ..... + ")
		mntl.WriteAttributes(minigo.InversionFond)
		mntl.Send(minigo.EncodeMessage("ENVOI"))
		mntl.WriteAttributes(minigo.FondNormal)

		mntl.WriteStringXY(1, 24, "CHOIX SERVICE ")
		mntl.WriteAttributes(minigo.InversionFond)
		mntl.Send(minigo.EncodeMessage("SOMMAIRE"))
		mntl.WriteAttributes(minigo.FondNormal)

		inputs.AppendInput("code_postal", minigo.NewInput(mntl, 14, 4, 5, 1, "", true))
		inputs.ActivateFirst()

		return minigo.NoOp
	})

	codePostalPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if len(inputs.ValueActive()) != 0 {
			infoLog.Printf("chosen code postal: %s\n", inputs.ValueActive())
			return inputs.ToMap(), minigo.QuitOp
		}
		warnLog.Println("empty code postal")
		return nil, minigo.NoOp
	})

	codePostalPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	codePostalPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key uint) {
		inputs.AppendKeyActive(byte(key))
	})

	codePostalPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	return codePostalPage
}
