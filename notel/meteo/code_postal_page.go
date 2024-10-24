package meteo

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/logs"
)

func NewCodePostalPage(mntl *minigo.Minitel) *minigo.Page {
	codePostalPage := minigo.NewPage("code_postal", mntl, nil)

	codePostalPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.Reset()

		mntl.SendVDT("static/meteo.vdt")
		mntl.ModeG0()

		mntl.MoveAt(9, 1)
		mntl.PrintAttributes("Prévisions Météo", minigo.DoubleHauteur)

		mntl.ReturnCol(2, 1)
		mntl.Helper("CODE POSTAL:       →", "ENVOI", minigo.FondVert, minigo.CaractereNoir)
		inputs.AppendInput("code_postal", minigo.NewInput(mntl, 11, 14, 5, 1, true))

		mntl.ReturnCol(4, 1)
		mntl.PrintAttributes("Observations en Direct", minigo.DoubleHauteur)

		mntl.ReturnCol(2, 1)
		mntl.Print("(Parfois panne Météo France...)")

		mntl.ReturnCol(1, 1)
		mntl.Print("Avec variations sur 24h")

		mntl.ReturnCol(2, 1)
		mntl.Helper("Consulter →", "SUITE", minigo.FondRouge, minigo.CaractereBlanc)

		mntl.ReturnCol(4, 1)
		mntl.Helper("Menu NOTEL", "SOMMAIRE", minigo.FondBleu, minigo.CaractereBlanc)

		inputs.InitAll()
		return minigo.NoOp
	})

	codePostalPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if len(inputs.ValueActive()) != 0 {
			logs.InfoLog("chosen code postal: %s\n", inputs.ValueActive())
			return inputs.ToMap(), minigo.QuitOp
		}

		logs.WarnLog("empty code postal\n")
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
		return nil, minigo.SommaireOp
	})

	codePostalPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SuiteOp
	})

	return codePostalPage
}
