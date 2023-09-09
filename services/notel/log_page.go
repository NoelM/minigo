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

func initLog(mntl *minigo.Minitel, form *minigo.Form, initData map[string]string) {
	mntl.CleanScreen()

	mntl.WriteAttributes(minigo.DoubleGrandeur, minigo.InversionFond)
	mntl.WriteStringXY(10, 10, "MINI-CHAT")

	mntl.WriteAttributes(minigo.GrandeurNormale, minigo.FondNormal)
	mntl.WriteStringXY(10, 12, "PSEUDO : ")
	mntl.CursorOnXY(10, 13)

	form.AppendInput("nick", minigo.NewInput(mntl, 10, 13, 10, 1, "", true))
	form.ActivateFirst()
}

func envoiLog(mntl *minigo.Minitel, form *minigo.Form) (map[string]string, int) {
	if len(form.ValueActive()) == 0 {
		warnLog.Println("empty nick input")
		return nil, minigo.QuitOp
	}
	mntl.Reset()

	infoLog.Printf("logged as: %s\n", form.ValueActive())
	return form.ToMap(), minigo.NoOp
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