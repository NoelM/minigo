package main

import (
	"github.com/NoelM/minigo"
)

func NewLogPage(mntl *minigo.Minitel) *minigo.Page {
	logPage := minigo.NewPage("log", mntl, nil)

	logPage.SetInitFunc(initLog)
	logPage.SetCharFunc(keyLog)
	logPage.SetEnvoiFunc(envoiLog)
	logPage.SetCorrectionFunc(correctionLog)
	logPage.SetSommaireFunc(sommaireLog)

	return logPage
}

func initLog(mntl *minigo.Minitel, form *minigo.Form, initData map[string]string) int {
	mntl.CleanScreen()

	mntl.WriteAttributes(minigo.DoubleGrandeur, minigo.InversionFond)
	mntl.WriteStringAt(10, 10, "MINI-CHAT")

	mntl.WriteAttributes(minigo.GrandeurNormale, minigo.FondNormal)
	mntl.WriteHelperAt(10, 13, "PSEUDO : ......... +", "ENVOI")

	mntl.WriteStringCenter(16, "En simultané sur libera.chat#minitel")

	form.AppendInput("nick", minigo.NewInput(mntl, 19, 13, 10, 1, "", true))
	form.ActivateFirst()

	return minigo.NoOp
}

func envoiLog(mntl *minigo.Minitel, form *minigo.Form) (map[string]string, int) {
	if len(form.ValueActive()) == 0 {
		warnLog.Println("empty nick input")
		return nil, minigo.QuitOp
	}
	mntl.Reset()

	infoLog.Printf("logged as: %s\n", form.ValueActive())
	return form.ToMap(), minigo.QuitOp
}

func sommaireLog(mntl *minigo.Minitel, form *minigo.Form) (map[string]string, int) {
	return nil, sommaireId
}

func correctionLog(mntl *minigo.Minitel, form *minigo.Form) (map[string]string, int) {
	form.CorrectionActive()
	return nil, minigo.NoOp
}

func keyLog(mntl *minigo.Minitel, form *minigo.Form, key uint) {
	form.AppendKeyActive(byte(key))
}
