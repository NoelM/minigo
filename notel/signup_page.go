package main

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/logs"
)

func NewSignUpPage(mntl *minigo.Minitel) *minigo.Page {
	signUpPage := minigo.NewPage("signup", mntl, nil)

	signUpPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.Reset()
		mntl.CleanScreen()
		mntl.SendVDT("static/connect.vdt")
		mntl.ModeG0()

		mntl.MoveAt(12, 1)
		mntl.Attributes(minigo.FondNormal, minigo.DoubleHauteur)
		mntl.Print("Inscrivez vous !")

		mntl.Attributes(minigo.GrandeurNormale)

		mntl.ReturnCol(3, 1)
		mntl.Print("PSEUDO:")
		inputs.AppendInput("login", minigo.NewInput(mntl, 15, 15, 10, 1, true))

		mntl.Right(18)
		mntl.Print("+")
		mntl.Right(1)
		mntl.Button("SUITE", minigo.FondBleu, minigo.CaractereBlanc)

		mntl.ReturnCol(2, 1)
		mntl.Print("MOT DE PASSE:")
		inputs.AppendInput("pwd", minigo.NewInput(mntl, 17, 15, 10, 1, true))

		mntl.Right(12)
		mntl.Print("+")
		mntl.Right(1)
		mntl.Button("SUITE", minigo.FondBleu, minigo.CaractereBlanc)

		mntl.ReturnCol(1, 1)
		mntl.Print("CONFIRMEZ:")
		inputs.AppendInput("pwdRepeat", minigo.NewInput(mntl, 18, 15, 10, 1, true))

		mntl.ReturnCol(2, 1)
		mntl.Helper("Validez →", "ENVOI", minigo.FondJaune, minigo.CaractereNoir)

		mntl.ReturnCol(3, 1)
		mntl.Print("Compte supprimé après 30j")

		mntl.ReturnCol(1, 3)
		mntl.Print("sans connexion")

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
	mntl.MoveAt(11, 1)
	mntl.CleanLine()
	mntl.PrintAttributesAt(11, 1, errorMsg, minigo.InversionFond)
}
