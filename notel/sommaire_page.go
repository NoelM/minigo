package main

import (
	"fmt"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/logs"
	"github.com/NoelM/minigo/notel/minichat"
	"github.com/NoelM/minigo/notel/profil"
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
			op = profil.ProfilService(m, UsersDb, nick)
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

	list := minigo.NewList(mntl, 8, 1, 17, 2)
	list.AppendItem(chatKey, "MINICHAT")
	list.AppendItem(meteoKey, "METEO")
	list.AppendItem(infoKey, "INFOS")
	list.AppendItem(sudokuKey, "SUDOKU")
	list.AppendItem(serveurKey, "SERVEUR")
	list.AppendItem(profilKey, "PROFIL")
	list.Display()

	mntl.MoveAt(19, 0)
	mntl.WriteAttributes(minigo.DoubleHauteur)
	mntl.WriteStringCenter("Prochaine soirée chat ?")

	mntl.WriteAttributes(minigo.GrandeurNormale)

	mntl.Return(1)
	mntl.WriteStringCenter("A vous de choisir")

	mntl.ReturnCol(4, 1)
	cntd := NbConnectedUsers.Load()
	if cntd < 2 {
		mntl.WriteString(fmt.Sprintf("> Connecté: %d", cntd))
	} else {
		mntl.WriteString(fmt.Sprintf("> Connectés: %d", cntd))
	}

	mntl.PrintHelperRight("CODE .... +", "ENVOI", minigo.FondBleu, minigo.CaractereBlanc)
	form.AppendInput("choice", minigo.NewInput(mntl, 24, 25, 4, 1, true))

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
