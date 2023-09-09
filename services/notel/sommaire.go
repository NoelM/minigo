package main

import (
	"strconv"

	"github.com/NoelM/minigo"
)

const noopId = -2
const quitId = -1

const (
	sommaireId = iota
	ircId
	meteoId
	guideId
)

func NewPageSommaire(mntl *minigo.Minitel) *minigo.Page {
	sommairePage := minigo.NewPage("sommaire", mntl, nil)

	sommairePage.SetInitFunc(initLog)
	sommairePage.SetCharFunc(keyLog)
	sommairePage.SetEnvoiFunc(envoiLog)
	sommairePage.SetCorrectionFunc(correctionLog)

	return sommairePage
}

func initSommaire(mntl *minigo.Minitel, form *minigo.Form, initData map[string]string) {
	mntl.CleanScreen()
	mntl.SendVDT("static/notel.vdt")

	list := minigo.NewList(mntl, []string{"Mini-Chat"})
	list.Display()

	form.AppendInput("choice", minigo.NewInput(mntl, 32, 24, 2, 1, "", true))
	form.ActivateFirst()
}

func envoiSommaire(mntl *minigo.Minitel, form *minigo.Form) (map[string]string, int) {
	if len(form.ValueActive()) == 0 {
		warnLog.Println("empty choice")
		return nil, minigo.NoOp
	}

	mntl.Reset()
	infoLog.Printf("chosen service: %s\n", form.ValueActive())

	id, err := strconv.Atoi(string(form.ValueActive()))
	if err != nil {
		warnLog.Println("unable to parse choice")
		return nil, minigo.NoOp
	}
	return form.ToMap(), id
}

func correctionSommaire(mntl *minigo.Minitel, form *minigo.Form) (map[string]string, int) {
	form.CorrectionActive()
	return nil, minigo.NoOp
}

func keySommaire(mntl *minigo.Minitel, form *minigo.Form, key uint) {
	form.AppendKeyActive(byte(key))
}
