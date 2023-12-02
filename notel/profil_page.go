package main

import (
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func RunPageProfil(mntl *minigo.Minitel, userDB *databases.UsersDatabase, pseudo string) (op int) {
	profilPage := minigo.NewPage("pioupiou:profil", mntl, nil)

	profilPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringLeft(2, "Mon compte")
		mntl.WriteAttributes(minigo.GrandeurNormale)

		mntl.HLine(3, 1, 40, minigo.HCenter)

		mntl.WriteStringLeft(5, "PSEUDO: "+pseudo)

		mntl.WriteStringLeft(7, "CHANGER MOT DE PASSE")
		mntl.WriteStringLeft(9, "Actuel:")
		inputs.AppendInput("now", minigo.NewInput(mntl, 9, 10, 10, 1, true))
		mntl.WriteStringLeft(10, "Nouveau:")
		inputs.AppendInput("new", minigo.NewInput(mntl, 10, 10, 10, 1, true))
		mntl.WriteStringLeft(11, "Nouveau:")
		inputs.AppendInput("newRep", minigo.NewInput(mntl, 11, 10, 10, 1, true))

		mntl.WriteHelperLeft(24, "Validez avec", "ENVOI")

		inputs.InitAll()
		return minigo.NoOp
	})

	profilPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key rune) {
		inputs.AppendKeyActive(key)
	})

	profilPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.ActivatePrev()
		return nil, minigo.NoOp
	})

	profilPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.ActivateNext()
		return nil, minigo.NoOp
	})

	profilPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	profilPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	profilPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		pwd := inputs.ToMap()

		if !userDB.LogUser(pseudo, pwd["now"]) {
			printPseudoMsg(mntl, inputs, "Erreur: mot de passe invalide !")
			return nil, minigo.NoOp
		}
		delete(pwd, "now")

		if pwd["new"] != pwd["newRep"] {
			printPseudoMsg(mntl, inputs, "Erreur: mdp. ne correspondent pas")
			return nil, minigo.NoOp
		}
		delete(pwd, "newRep")

		if !userDB.ChangePassword(pseudo, pwd["new"]) {
			printPseudoMsg(mntl, inputs, "Erreur interne")
			return nil, minigo.NoOp
		}

		printPseudoMsg(mntl, inputs, "Mot de passe modifié !")
		return nil, minigo.NoOp
	})

	_, op = profilPage.Run()
	return op
}

func printPseudoMsg(mntl *minigo.Minitel, inputs *minigo.Form, msg string) {
	mntl.WriteStringLeft(12, msg)
	time.Sleep(2 * time.Second)
	mntl.CleanLine()

	inputs.ResetAll()
	inputs.ActivateFirst()
}
