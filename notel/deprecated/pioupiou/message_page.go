package pioupiou

import "github.com/NoelM/minigo"

func NewPageMessage(mntl *minigo.Minitel, pseudo string) *minigo.Page {
	homePage := minigo.NewPage("pioupiou:message", mntl, nil)

	homePage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		mntl.PrintAttributesAt(10, 1, "Publiez un message", minigo.DoubleHauteur)
		inputs.AppendInput("message", minigo.NewInput(mntl, 12, 1, 40, 5, true))

		mntl.HelperRightAt(18, "Publiez avec", "ENVOI")
		mntl.HelperLeftAt(24, "Menu", "SOMMAIRE")

		inputs.InitAll()
		return minigo.NoOp
	})

	homePage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		msg := inputs.ValueActive()
		if len(msg) == 0 {
			return nil, minigo.NoOp
		}
		PiouPiou.Post(pseudo, string(msg))

		return nil, minigo.SommaireOp
	})

	homePage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	homePage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key int32) {
		inputs.AppendKeyActive(key)
	})

	homePage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	return homePage
}
