package blog

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/logs"
)

func ListPage(m *minigo.Minitel, articles []Article) *minigo.Page {
	listPage := minigo.NewPage("blog:list", m, nil)

	listPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		m.Reset()
		m.CursorOff()
		m.SendVDT("static/blog.vdt")

		// Display header
		m.ModeG0()
		m.Attributes(minigo.FondNoir, minigo.CaractereBlanc, minigo.GrandeurNormale)

		mntl.MoveAt(7, 1)
		mntl.Attributes(minigo.DoubleHauteur)
		mntl.Print("Articles")
		mntl.Attributes(minigo.GrandeurNormale)

		// Display available channels
		names := make([]string, 0, len(articles))
		for _, article := range articles {
			names = append(names, article.Title)
		}
		list := minigo.NewListEnum(mntl, names, 9, 1, 22, 1)
		list.Display()

		mntl.MoveAt(24, 0)
		mntl.HelperRight("ARTICLE: .. +", "ENVOI", minigo.FondVert, minigo.CaractereNoir)
		inputs.AppendInput("article", minigo.NewInput(mntl, 24, 27, 2, 1, true))

		inputs.InitAll()
		return minigo.NoOp
	})

	listPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		logs.InfoLog("selected article: %s\n", inputs.ValueActive())
		return map[string]string{"article": inputs.ValueActive()}, minigo.EnvoiOp
	})

	listPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	listPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	listPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key int32) {
		inputs.AppendKeyActive(key)
	})

	return listPage
}
