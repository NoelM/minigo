package main

import (
	"fmt"
	"strconv"

	"github.com/NoelM/minigo"
)

const (
	sommaireId = iota
	chatId
	meteoId
	infoId
	serveurId
)

func SommaireHandler(m *minigo.Minitel, login string) {
	infoLog.Println("enters sommaire handler")
	var id int
	for id >= sommaireId {
		switch id {
		case sommaireId:
			_, id = NewPageSommaire(m).Run()
		case chatId:
			id = ServiceMiniChat(m, login)
		case meteoId:
			id = ServiceMeteo(m)
		case infoId:
			_, id = NewPageInfo(m).Run()
		case serveurId:
			_, id = NewServeurPage(m).Run()
		default:
			id = sommaireId
		}
	}
	infoLog.Println("quits sommaire handler")
}

func NewPageSommaire(mntl *minigo.Minitel) *minigo.Page {
	sommairePage := minigo.NewPage("sommaire", mntl, nil)

	sommairePage.SetInitFunc(initSommaire)
	sommairePage.SetCharFunc(keySommaire)
	sommairePage.SetEnvoiFunc(envoiSommaire)
	sommairePage.SetCorrectionFunc(correctionSommaire)

	return sommairePage
}

func initSommaire(mntl *minigo.Minitel, form *minigo.Form, initData map[string]string) int {
	mntl.CleanScreen()
	mntl.SendVDT("static/notel.vdt")
	mntl.WriteAttributes(minigo.FondNormal, minigo.GrandeurNormale)

	list := minigo.NewList(mntl, []string{"MINICHAT", "METEO", "INFOS", "SERVEUR"})
	list.Display()

	mntl.WriteStringCenter(18, "Le serveur est multivoies !")
	mntl.WriteStringCenter(19, "RDV Dim. 12 Nov. à 20h sur le chat")

	mntl.WriteStringLeft(24, fmt.Sprintf("> Connectés: %d", NbConnectedUsers.Load()))
	form.AppendInput("choice", minigo.NewInput(mntl, 24, 32, 2, 1, true))
	form.InitAll()

	return minigo.NoOp
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
		return nil, sommaireId
	}
	return form.ToMap(), id
}

func correctionSommaire(mntl *minigo.Minitel, form *minigo.Form) (map[string]string, int) {
	form.CorrectionActive()
	return nil, minigo.NoOp
}

func keySommaire(mntl *minigo.Minitel, form *minigo.Form, key rune) {
	form.AppendKeyActive(key)
}
