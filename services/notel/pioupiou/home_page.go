package pioupiou

import (
	"github.com/NoelM/minigo"
)

func NewPageHome(mntl *minigo.Minitel) *minigo.Page {
	homePage := minigo.NewPage("pioupiou:home", mntl, nil)

	homePage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		mntl.WriteStringAtWithAttributes(10, 1, "Microblog sur Minitel", minigo.DoubleHauteur)
		mntl.WriteHelperLeft(11, "Retrouvez cet aide avec", "GUIDE")

		mntl.WriteStringLeft(13, "/FIL  Ouvrir son FIL personnel")
		mntl.WriteStringLeft(14, "/MSG  Ecrivez un MESSAGE")
		mntl.WriteStringLeft(15, "/ANU  ANNUAIRE des profils")
		//mntl.WriteStringLeft(15, "/NOT  Ouvrir ses NOTIFICATIONS")
		//mntl.WriteStringLeft(16, "/CRC  CHERCHEZ profil ou mot-dièse")
		//mntl.WriteStringLeft(17, "/PRO  Ouvrir son PROFIL")

		//mntl.WriteStringLeft(19, "Mentionnez un utilisateur avec @PSEUDO")
		//mntl.WriteStringLeft(20, "Utilisez des mots-dièses #EXEMPLE")

		mntl.WriteHelperLeft(24, "Menu NOTEL", "SOMMAIRE")
		inputs.AppendInput("command", minigo.NewInput(mntl, 24, 29, 4, 1, true))
		mntl.WriteHelperRight(24, ".... +", "ENVOI")

		inputs.InitAll()
		return minigo.NoOp
	})

	homePage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		cmdString := inputs.ValueActive()
		if len(cmdString) == 0 {
			return nil, minigo.NoOp
		}

		cmdId := ParseCommandString(string(cmdString))
		if cmdId < 0 {
			PrintErrorMessage("Commande inconnue, utilisez GUIDE")
		}
		return nil, cmdId
	})

	homePage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	homePage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key uint) {
		inputs.AppendKeyActive(byte(key))
	})

	homePage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	return homePage
}
