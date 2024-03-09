package profil

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func RunSupprPage(mntl *minigo.Minitel, userDB *databases.UsersDatabase, pseudo string) (op int) {
	supprPage := minigo.NewPage("suppr", mntl, nil)

	supprPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		mntl.MoveAt(2, 0)
		mntl.WriteStringWithAttributes("Supprimer son Compte", minigo.DoubleHauteur)

		mntl.Return(1)                 // Row 3
		mntl.HLine(40, minigo.HCenter) // Returns Row 4

		mntl.Return(1) // Row 5
		mntl.WriteString("Ecrire \"supprimer\":")
		inputs.AppendInput("suppr", minigo.NewInput(mntl, 5, 22, 9, 1, true))

		mntl.Return(2) // Row 7
		mntl.WriteString("Mot de passe:")
		inputs.AppendInput("pwd", minigo.NewInput(mntl, 7, 22, 10, 1, true))

		mntl.Return(2) // Row 9
		mntl.PrintHelper("Validez avec", "ENVOI", minigo.FondVert, minigo.CaractereNoir)

		inputs.InitAll()
		return minigo.NoOp
	})

	supprPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key rune) {
		inputs.AppendKeyActive(key)
	})

	supprPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.ActivatePrev()
		return nil, minigo.NoOp
	})

	supprPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.ActivateNext()
		return nil, minigo.NoOp
	})

	supprPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	supprPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	supprPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		data := inputs.ToMap()

		if !userDB.LogUser(pseudo, data["pwd"]) {
			printErrorMsg(mntl, inputs, "Erreur: mot de passe invalide !")
			return nil, minigo.NoOp
		}
		delete(data, "pwd")

		if err := userDB.DeleteUser(pseudo); err != nil {
			printErrorMsg(mntl, inputs, "Erreur interne")
			return nil, minigo.NoOp
		}

		printErrorMsg(mntl, inputs, "Compte supprimé, au revoir !")
		mntl.In <- minigo.ConnexionFin
		return nil, minigo.NoOp
	})

	_, op = supprPage.Run()
	return op
}
