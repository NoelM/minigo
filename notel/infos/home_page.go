package infos

import (
	"strconv"

	"github.com/NoelM/minigo"
)

func NewHomePage(m *minigo.Minitel) *minigo.Page {
	home := minigo.NewPage("infos:home", m, nil)

	names := []string{
		"France 24",
		"franceinfo",
		"Le Monde (en continu)",
		"Le Monde (à la une)",
		"Libération",
		"BBC News",
		"The Verge",
	}

	feeds := []string{
		France24Rss,
		FranceInfoRss,
		LeMondeLiveRss,
		LeMondeRss,
		LiberationRss,
		BBCRss,
		TheVergeRss,
	}

	home.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.Reset()
		mntl.SendVDT("static/infos.vdt")

		mntl.ModeG0()
		list := minigo.NewListEnum(mntl, names, 6, 1, 20, 2)
		list.Display()

		mntl.MoveAt(24, 0)
		mntl.HelperRight("Numéro   +", "ENVOI", minigo.FondBleu, minigo.CaractereBlanc)

		inputs.AppendInput("num", minigo.NewInput(mntl, 24, 28, 1, 1, true))
		inputs.InitAll()

		return minigo.NoOp
	})

	home.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	home.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key rune) {
		inputs.AppendKeyActive(key)
	})

	home.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		val := inputs.ValueActive()

		servId, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			inputs.ResetAll()
			return nil, minigo.NoOp
		}

		if servId < 0 || servId > 7 {
			inputs.ResetAll()
			return nil, minigo.NoOp
		}

		return map[string]string{
				"name": names[servId-1],
				"url":  feeds[servId-1],
			},
			minigo.EnvoiOp
	})

	return home
}
