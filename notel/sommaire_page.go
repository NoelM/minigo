package main

import (
	"fmt"
	"strings"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/annuaire"
	"github.com/NoelM/minigo/notel/infos"
	"github.com/NoelM/minigo/notel/logs"
	"github.com/NoelM/minigo/notel/meteo"
	"github.com/NoelM/minigo/notel/minichat"
	"github.com/NoelM/minigo/notel/profil"
	"github.com/NoelM/minigo/notel/serveur"
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
	annuaireId
)

const (
	chatKey     = "*CHA"
	meteoKey    = "*MTO"
	infoKey     = "*INF"
	serveurKey  = "*SRV"
	sudokuKey   = "*SDK"
	profilKey   = "*PRO"
	annuaireKey = "*ANU"
)

var ServIdMap = map[string]int{
	chatKey:     chatId,
	meteoKey:    meteoId,
	infoKey:     infoId,
	serveurKey:  serveurId,
	sudokuKey:   sudokuId,
	profilKey:   profilId,
	annuaireKey: annuaireId,
}

func SommaireHandler(m *minigo.Minitel, nick string, metrics *Metrics) {
	logs.InfoLog("enters sommaire handler\n")

	var op int
	var choice map[string]string

	for op != minigo.DisconnectOp {
		choice, op = NewPageSommaire(m, metrics).Run()
		serviceId, ok := ServIdMap[strings.ToUpper(choice["choice"])]
		if !ok {
			continue
		}

		switch serviceId {
		case chatId:
			op = minichat.RunChatPage(m, MessageDb, &metrics.ConnectedUsers, nick, metrics.MessagesCount)
		case meteoId:
			op = meteo.MeteoService(m, CommuneDb)
		case infoId:
			op = infos.ServiceInfo(m)
		case serveurId:
			_, op = serveur.NewServeurPage(m).Run()
		case sudokuId:
			op = sudoku.SudokuService(m, nick)
		case profilId:
			op = profil.ProfilService(m, UsersDb, nick)
		case annuaireId:
			op = annuaire.AnnuaireService(m, UsersDb)
		}
	}
	logs.InfoLog("quits sommaire handler\n")
}

func NewPageSommaire(mntl *minigo.Minitel, metrics *Metrics) *minigo.Page {
	sommairePage := minigo.NewPage("sommaire", mntl, nil)

	sommairePage.SetInitFunc(func(mntl *minigo.Minitel, form *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.SendVDT("static/notel.vdt")

		mntl.ModeG0()
		mntl.Attributes(minigo.FondNoir, minigo.CaractereBlanc, minigo.GrandeurNormale)

		list := minigo.NewList(mntl, 8, 1, 17, 2)
		list.AppendItem(chatKey, "MINICHAT")
		list.AppendItem(meteoKey, "METEO")
		list.AppendItem(infoKey, "INFOS")
		list.AppendItem(sudokuKey, "SUDOKU")
		list.AppendItem(serveurKey, "SERVEUR")
		list.AppendItem(profilKey, "PROFIL")
		list.AppendItem(annuaireKey, "ANNUAIRE")
		list.Display()

		mntl.MoveAt(19, 0)
		mntl.Attributes(minigo.DoubleHauteur)
		mntl.PrintCenter("NOTEL est de retour !")

		mntl.Attributes(minigo.GrandeurNormale)

		mntl.Return(1)
		mntl.PrintCenter("Bienvenue sur le Minitel")

		mntl.ReturnCol(4, 1)
		cntd := metrics.ConnectedUsers.Load()
		if cntd < 2 {
			mntl.Print(fmt.Sprintf("> Connecté: %d", cntd))
		} else {
			mntl.Print(fmt.Sprintf("> Connectés: %d", cntd))
		}

		mntl.HelperRight("CODE .... +", "ENVOI", minigo.FondBleu, minigo.CaractereBlanc)
		form.AppendInput("choice", minigo.NewInput(mntl, 24, 25, 4, 1, true))

		form.InitAll()

		return minigo.NoOp
	})

	sommairePage.SetCharFunc(keySommaire)
	sommairePage.SetEnvoiFunc(envoiSommaire)
	sommairePage.SetCorrectionFunc(correctionSommaire)

	return sommairePage
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
