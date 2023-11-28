package main

import (
	"fmt"

	"github.com/NoelM/minigo"
)

const (
	sommaireId = iota
	chatId
	meteoId
	infoId
	serveurId
	sudokuId
)

const (
	chatKey    = "*CHA"
	meteoKey   = "*MTO"
	infoKey    = "*INF"
	serveurKey = "*SRV"
	sudokuKey  = "*SDK"
)

var ServIdMap = map[string]int{
	chatKey:    chatId,
	meteoKey:   meteoId,
	infoKey:    infoId,
	serveurKey: serveurId,
	sudokuKey:  sudokuId,
}

func SommaireHandler(m *minigo.Minitel, login string) {
	infoLog.Println("enters sommaire handler")

	var op int
	var choice map[string]string

	for op != minigo.DisconnectOp {
		choice, op = NewPageSommaire(m).Run()
		serviceId, ok := ServIdMap[choice["choice"]]
		if !ok {
			continue
		}

		switch serviceId {
		case chatId:
			op = ServiceMiniChat(m, login)
		case meteoId:
			op = ServiceMeteo(m)
		case infoId:
			_, op = NewPageInfo(m).Run()
		case serveurId:
			_, op = NewServeurPage(m).Run()
		case sudokuId:
			op = SudokuService(m)
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
	sommairePage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, chatId
	})

	return sommairePage
}

func initSommaire(mntl *minigo.Minitel, form *minigo.Form, initData map[string]string) int {
	mntl.CleanScreen()
	mntl.SendVDT("static/notel.vdt")
	mntl.WriteAttributes(minigo.FondNormal, minigo.GrandeurNormale)
	mntl.ModeG0()

	list := minigo.NewList(mntl, 8, 1, 20, 2)
	list.AppendItem(chatKey, "MINICHAT")
	list.AppendItem(meteoKey, "METEO")
	list.AppendItem(infoKey, "INFOS")
	list.AppendItem(sudokuKey, "SUDOKU")
	list.AppendItem(serveurKey, "SERVEUR")

	list.Display()

	mntl.WriteAttributes(minigo.Clignotement, minigo.DoubleHauteur)
	mntl.WriteStringCenter(19, "→ Rendez-vous ←")
	mntl.WriteAttributes(minigo.Fixe, minigo.GrandeurNormale)

	mntl.WriteStringCenter(20, "Dimanche 3 Déc. à 20h")

	mntl.WriteStringLeft(24, fmt.Sprintf("> Connectés: %d", NbConnectedUsers.Load()))
	mntl.WriteHelperRight(24, "CHOIX ....", "ENVOI")

	form.AppendInput("choice", minigo.NewInput(mntl, 24, 30, 4, 1, true))
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

	return form.ToMap(), minigo.SommaireOp
}

func correctionSommaire(mntl *minigo.Minitel, form *minigo.Form) (map[string]string, int) {
	form.CorrectionActive()
	return nil, minigo.NoOp
}

func keySommaire(mntl *minigo.Minitel, form *minigo.Form, key rune) {
	form.AppendKeyActive(key)
}
