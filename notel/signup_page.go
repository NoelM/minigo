package main

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/logs"
)

func NewSignUpPage(mntl *minigo.Minitel) *minigo.Page {
	signUpPage := minigo.NewPage("signup", mntl, nil)

	signUpPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.SendVDT("static/connect.vdt")
		mntl.ModeG0()

		mntl.WriteStringAtWithAttributes(10, 1, "Inscrivez vous !", minigo.FondNormal, minigo.DoubleHauteur)
		mntl.WriteStringLeft(11, "Compte supprimé si")
		mntl.WriteStringLeft(12, "30j sans connexion")

		mntl.WriteStringLeft(15, "PSEUDO:")
		inputs.AppendInput("login", minigo.NewInput(mntl, 15, 15, 10, 1, true))
		mntl.WriteStringLeft(16, "MOT DE PASSE:")
		inputs.AppendInput("pwd", minigo.NewInput(mntl, 16, 15, 10, 1, true))
		mntl.WriteStringLeft(17, "MOT DE PASSE:")
		inputs.AppendInput("pwdRepeat", minigo.NewInput(mntl, 17, 15, 10, 1, true))

		mntl.WriteHelperLeft(19, "Validez avec", "ENVOI")

		inputs.InitAll()
		return minigo.NoOp
	})

	signUpPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	signUpPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key int32) {
		inputs.AppendKeyActive(key)
	})

	signUpPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		creds := inputs.ToMap()
		inputs.ResetAll()

		if len(creds["login"]) == 0 || len(creds["pwd"]) == 0 {
			printSignUpError(mntl, "Pseudo ou MDP vides")
			return nil, minigo.NoOp
		}

		if creds["pwd"] != creds["pwdRepeat"] {
			printSignUpError(mntl, "Mots de passes non indentiques")
			return nil, minigo.NoOp
		}

		if UsersDb.UserExists(creds["login"]) {
			printSignUpError(mntl, "Pseudo déjà utilisé")
			return nil, minigo.NoOp
		}

		err := UsersDb.AddUser(creds["login"], creds["pwd"])
		delete(creds, "pwd")
		delete(creds, "pwdRepeat")

		if err == nil {
			logs.InfoLog("new signup for user=%s\n", creds["login"])
			return creds, minigo.EnvoiOp
		} else {
			printSignUpError(mntl, "Erreur serveur")
			return nil, minigo.NoOp
		}
	})

	signUpPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	signUpPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.ActivateNext()
		return nil, minigo.NoOp
	})

	signUpPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.ActivatePrev()
		return nil, minigo.NoOp
	})
	return signUpPage
}

func printSignUpError(mntl *minigo.Minitel, errorMsg string) {
	mntl.MoveCursorAt(11, 1)
	mntl.CleanLine()
	mntl.WriteStringAtWithAttributes(11, 1, errorMsg, minigo.InversionFond)
}
