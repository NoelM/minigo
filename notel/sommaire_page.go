package main

import (
	"fmt"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/logs"
	"github.com/NoelM/minigo/notel/minichat"
	"github.com/NoelM/minigo/notel/sudoku"
)

const (
	sommaireId = iota
	chatId
	meteoId
	infoId
	serveurId
	sudokuId
	profilId
)

const (
	chatKey    = "*CHA"
	meteoKey   = "*MTO"
	infoKey    = "*INF"
	serveurKey = "*SRV"
	sudokuKey  = "*SDK"
	profilKey  = "*PRO"
)

var ServIdMap = map[string]int{
	chatKey:    chatId,
	meteoKey:   meteoId,
	infoKey:    infoId,
	serveurKey: serveurId,
	sudokuKey:  sudokuId,
	profilKey:  profilId,
}

func SommaireHandler(m *minigo.Minitel, nick string) {
	logs.InfoLog("enters sommaire handler\n")

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
			op = minichat.RunChatPage(m, MessageDb, &NbConnectedUsers, nick, promMsgNb)
		case meteoId:
			op = ServiceMeteo(m)
		case infoId:
			_, op = NewPageInfo(m).Run()
		case serveurId:
			_, op = NewServeurPage(m).Run()
		case sudokuId:
			op = sudoku.SudokuService(m, nick)
		case profilId:
			op = RunPageProfil(m, UsersDb, nick)
		}
	}
	logs.InfoLog("quits sommaire handler\n")
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

	mntl.ModeG0()
	mntl.WriteAttributes(minigo.FondNoir, minigo.CaractereBlanc, minigo.GrandeurNormale)

	listLeft := minigo.NewList(mntl, 8, 1, 20, 2)
	listLeft.AppendItem(chatKey, "MINICHAT")
	listLeft.AppendItem(meteoKey, "METEO")
	listLeft.AppendItem(infoKey, "INFOS")
	listLeft.AppendItem(sudokuKey, "SUDOKU")
	listLeft.AppendItem(serveurKey, "SERVEUR")
	listLeft.Display()

	listRight := minigo.NewList(mntl, 8, 20, 20, 2)
	listRight.AppendItem(profilKey, "PROFIL")
	listRight.Display()

	mntl.WriteAttributes(minigo.DoubleHauteur)
	mntl.WriteStringCenterAt(19, "! NOTEL est de retour !")
	mntl.WriteAttributes(minigo.GrandeurNormale)

	mntl.WriteStringCenterAt(20, "RDV Dim. 3 Mars à 20h")

	cntd := NbConnectedUsers.Load()
	if cntd < 2 {
		mntl.WriteStringLeftAt(24, fmt.Sprintf("> Connecté: %d", cntd))
	} else {
		mntl.WriteStringLeftAt(24, fmt.Sprintf("> Connectés: %d", cntd))
	}

	mntl.WriteHelperRightAt(24, "CODE ....", "ENVOI")
	form.AppendInput("choice", minigo.NewInput(mntl, 24, 30, 4, 1, true))

	form.InitAll()

	return minigo.NoOp
}

func envoiSommaire(mntl *minigo.Minitel, form *minigo.Form) (map[string]string, int) {
	if len(form.ValueActive()) == 0 {
		logs.WarnLog("empty choice\n")
		return nil, minigo.NoOp
	}

	mntl.Reset()
	logs.InfoLog("chosen service: %s\n", form.ValueActive())

	return form.ToMap(), minigo.SommaireOp
}

func correctionSommaire(mntl *minigo.Minitel, form *minigo.Form) (map[string]string, int) {
	form.CorrectionActive()
	return nil, minigo.NoOp
}

func keySommaire(mntl *minigo.Minitel, form *minigo.Form, key rune) {
	form.AppendKeyActive(key)
}
