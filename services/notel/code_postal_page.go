package main

import (
	"github.com/NoelM/minigo"
)

func NewCodePostalPage(mntl *minigo.Minitel) *minigo.Page {
	codePostalPage := minigo.NewPage("code_postal", mntl, nil)

	codePostalPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		mntl.SendVDT("static/meteo.vdt")
		mntl.ModeG0()

		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringLeft(10, "Prévisions Météo")
		mntl.WriteAttributes(minigo.GrandeurNormale)

		mntl.WriteHelperLeft(12, "CODE POSTAL:       +", "ENVOI")
		inputs.AppendInput("code_postal", minigo.NewInput(mntl, 12, 14, 5, 1, true))

		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringLeft(16, "Observations en Direct")
		mntl.WriteAttributes(minigo.GrandeurNormale)
		mntl.WriteStringLeft(18, "Avec variations sur 24h")

		mntl.WriteHelperLeft(20, "APPUYEZ SUR", "SUITE")
		mntl.WriteHelperLeft(24, "Menu NOTEL", "SOMMAIRE")

		inputs.InitAll()
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

	codePostalPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key rune) {
		inputs.AppendKeyActive(key)
	})

	codePostalPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	codePostalPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SuiteOp
	})

	return codePostalPage
}
