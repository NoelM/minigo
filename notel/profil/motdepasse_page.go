package profil

import (
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func RunMDPPage(mntl *minigo.Minitel, userDB *databases.UsersDatabase, pseudo string) (op int) {
	mdpPage := minigo.NewPage("mdp", mntl, nil)

	mdpPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.Reset()

		mntl.MoveAt(2, 0)
		mntl.PrintAttributes("Changer Mot de Passe", minigo.DoubleHauteur)

		mntl.Return(1)                 // Row 3
		mntl.HLine(40, minigo.HCenter) // Returns Row 4

		mntl.Return(1) // Row 5
		mntl.Print("MDP Actuel:")
		inputs.AppendInput("now", minigo.NewInput(mntl, 5, 13, 10, 1, true))

		mntl.Return(2) // Row 7
		mntl.Print("Nouveau MDP:")
		inputs.AppendInput("new", minigo.NewInput(mntl, 7, 13, 10, 1, true))

		mntl.Return(1) // Row 8
		mntl.Print("Confirmez:")
		inputs.AppendInput("newRep", minigo.NewInput(mntl, 8, 13, 10, 1, true))

		mntl.Return(2) // Row 12
		mntl.Helper("Validez avec", "ENVOI", minigo.FondVert, minigo.CaractereNoir)

		inputs.InitAll()
		return minigo.NoOp
	})

	mdpPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key rune) {
		inputs.AppendKeyActive(key)
	})

	mdpPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.ActivatePrev()
		return nil, minigo.NoOp
	})

	mdpPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.ActivateNext()
		return nil, minigo.NoOp
	})

	mdpPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	mdpPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	mdpPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		pwd := inputs.ToMap()

		if !userDB.LogUser(pseudo, pwd["now"]) {
			printErrorMsg(mntl, inputs, "Erreur: mot de passe invalide !")
			return nil, minigo.NoOp
		}
		delete(pwd, "now")

		if pwd["new"] != pwd["newRep"] {
			printErrorMsg(mntl, inputs, "Erreur: mdp. ne correspondent pas")
			return nil, minigo.NoOp
		}
		delete(pwd, "newRep")

		if !userDB.ChangePassword(pseudo, pwd["new"]) {
			printErrorMsg(mntl, inputs, "Erreur interne")
			return nil, minigo.NoOp
		}

		printErrorMsg(mntl, inputs, "Mot de passe modifié !")
		return nil, minigo.SommaireOp
	})

	_, op = mdpPage.Run()
	return op
}

func printErrorMsg(mntl *minigo.Minitel, inputs *minigo.Form, msg string) {
	mntl.MoveAt(12, 0)
	mntl.PrintAttributes(msg, minigo.InversionFond)

	time.Sleep(2 * time.Second)
	mntl.CleanLine()

	inputs.ResetAll()
	inputs.ActivateFirst()
}
