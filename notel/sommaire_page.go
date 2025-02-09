package main

import (
	"fmt"
	"strings"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/annuaire"
	"github.com/NoelM/minigo/notel/blog"
	"github.com/NoelM/minigo/notel/infos"
	"github.com/NoelM/minigo/notel/logs"
	"github.com/NoelM/minigo/notel/meteo"
	"github.com/NoelM/minigo/notel/minichat"
	"github.com/NoelM/minigo/notel/profil"
	"github.com/NoelM/minigo/notel/stats"
	"github.com/NoelM/minigo/notel/superchat"
)

const (
	sommaireId = iota
	chatId
	superChatId
	meteoId
	infoId
	statsId
	profilId
	annuaireId
	blogId
)

const (
	chatKey      = "*CHA"
	superChatKey = "*SCA"
	meteoKey     = "*MTO"
	infoKey      = "*INF"
	statsKey     = "*STA"
	profilKey    = "*PRO"
	annuaireKey  = "*ANU"
	blogKey      = "*BLO"
)

var ServIdMap = map[string]int{
	chatKey:      chatId,
	superChatKey: superChatId,
	meteoKey:     meteoId,
	infoKey:      infoId,
	statsKey:     statsId,
	profilKey:    profilId,
	annuaireKey:  annuaireId,
	blogKey:      blogId,
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
		case superChatId:
			op = superchat.ServiceSuperchat(m, MessageDb, &metrics.ConnectedUsers, nick, metrics.MessagesCount)
		case meteoId:
			op = meteo.MeteoService(m, CommuneDb)
		case infoId:
			op = infos.ServiceInfo(m)
		case statsId:
			_, op = stats.NewStatsPage(m).Run()
		case blogId:
			op = blog.ServiceBlog(m, BlogDbPath)
		case profilId:
			op = profil.ProfilService(m, UsersDb, nick)
		case annuaireId:
			op = annuaire.AnnuaireService(m, UsersDb)
		}
	}
	logs.InfoLog("sommaire: quits handler\n")
}

func NewPageSommaire(mntl *minigo.Minitel, metrics *Metrics) *minigo.Page {
	sommairePage := minigo.NewPage("sommaire", mntl, nil)

	sommairePage.SetInitFunc(func(mntl *minigo.Minitel, form *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.CursorOff()
		mntl.SendVDT("static/notel.vdt")

		mntl.ModeG0()
		mntl.Attributes(minigo.FondNoir, minigo.CaractereBlanc, minigo.GrandeurNormale)

		list := minigo.NewList(mntl, 8, 1, 20, 2)
		list.AppendItem(chatKey, "MINICHAT")
		list.AppendItem(superChatKey, "SUPERCHAT (bêta)")
		list.AppendItem(meteoKey, "METEO")
		list.AppendItem(infoKey, "INFOS")
		list.AppendItem(blogKey, "BLOG")
		//list.AppendItem(statsKey, "STATS")
		list.AppendItem(profilKey, "PROFIL")
		list.AppendItem(annuaireKey, "ANNUAIRE")
		list.Display()

		mntl.MoveAt(24, 0)
		mntl.Attributes(minigo.FondBleu, minigo.CaractereBlanc)

		loggedCnt := metrics.CountLogged()
		if loggedCnt < 2 {
			// Whitespace required to activate the background
			mntl.Print(fmt.Sprintf(" Connecté: %d", loggedCnt))
		} else {
			// Whitespace required to activate the background
			mntl.Print(fmt.Sprintf(" Connectés: %d", loggedCnt))
		}

		mntl.HelperRight("CODE .... +", "ENVOI", minigo.FondVert, minigo.CaractereNoir)
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
		logs.WarnLog("sommaire: empty choice\n")
		return nil, minigo.NoOp
	}

	mntl.Reset()
	logs.InfoLog("sommaire: chosen service: %s\n", form.ValueActive())

	return form.ToMap(), minigo.SommaireOp
}

func correctionSommaire(mntl *minigo.Minitel, form *minigo.Form) (map[string]string, int) {
	form.CorrectionActive()
	return nil, minigo.NoOp
}

func keySommaire(mntl *minigo.Minitel, form *minigo.Form, key rune) {
	form.AppendKeyActive(key)
}
