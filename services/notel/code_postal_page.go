package main

import (
	"github.com/NoelM/minigo"
)

func NewCodePostalPage(mntl *minigo.Minitel) *minigo.Page {
	codePostalPage := minigo.NewPage("code_postal", mntl, nil)

	codePostalPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		mntl.SendVDT("static/meteo.vdt")

		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringCenter(6, "Prévisions Météo")
		mntl.WriteAttributes(minigo.GrandeurNormale)

		mntl.WriteHelperLeft(8, "CODE POSTAL: ..... +", "ENVOI")

		mntl.WriteHelperLeft(24, "CHOIX SERVICE", "SOMMAIRE")

		inputs.AppendInput("code_postal", minigo.NewInput(mntl, 14, 8, 5, 1, "", true))
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
