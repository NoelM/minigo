package main

import (
	"github.com/NoelM/minigo"
)

func NewCodePostalPage(mntl *minigo.Minitel) *minigo.Page {
	codePostalPage := minigo.NewPage("code_postal", mntl, nil)

	codePostalPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		mntl.SendVDT("static/meteo.vdt")
		mntl.Send([]byte{minigo.Si})

		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringLeft(10, "Prévisions Météo")
		mntl.WriteAttributes(minigo.GrandeurNormale)

		mntl.WriteHelperLeft(12, "CODE POSTAL: ..... +", "ENVOI")

		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringLeft(16, "Observations en Direct")
		mntl.WriteAttributes(minigo.GrandeurNormale)

		mntl.WriteHelperLeft(18, "APPUYEZ SUR", "SUITE")

		mntl.WriteHelperLeft(12, "CODE POSTAL: ..... +", "ENVOI")

		inputs.AppendInput("code_postal", minigo.NewInput(mntl, 14, 12, 5, 1, "", true))
		inputs.ActivateFirst()

		return minigo.NoOp
	})

	codePostalPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if len(inputs.ValueActive()) != 0 {
			infoLog.Printf("chosen code postal: %s\n", inputs.ValueActive())
			return inputs.ToMap(), minigo.QuitPageOp
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

	codePostalPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SuiteOp
	})

	return codePostalPage
}
