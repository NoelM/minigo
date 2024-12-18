package main

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/logs"
)

func NewPageSignIn(mntl *minigo.Minitel) *minigo.Page {
	signInPage := minigo.NewPage("signIn", mntl, nil)

	signInPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.SendVDT("static/connect.vdt")
		mntl.ModeG0()

		mntl.MoveAt(12, 1)
		mntl.Attributes(minigo.FondNormal, minigo.DoubleHauteur)
		mntl.Print("Connectez vous au serveur")

		mntl.Attributes(minigo.GrandeurNormale)

		mntl.ReturnCol(3, 1)
		mntl.Print("PSEUDO:")
		inputs.AppendInput("login", minigo.NewInput(mntl, 15, 15, 10, 1, true))

		mntl.Right(18)
		mntl.Print("+")
		mntl.Right(1)
		mntl.Button("SUITE", minigo.FondBleu, minigo.CaractereBlanc)

		mntl.ReturnCol(1, 1)
		mntl.Print("MOT DE PASSE:")
		inputs.AppendInput("pwd", minigo.NewInput(mntl, 16, 15, 10, 1, true))

		mntl.ReturnCol(2, 1)
		mntl.Helper("Validez →", "ENVOI", minigo.FondJaune, minigo.CaractereNoir)

		mntl.ReturnCol(6, 1)
		mntl.Helper("Première visite →", "GUIDE", minigo.FondVert, minigo.CaractereNoir)

		inputs.InitAll()

		mntl.CursorOn()
		return minigo.NoOp
	})

	signInPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key int32) {
		inputs.AppendKeyActive(key)
	})

	signInPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		creds := inputs.ToMap()
		inputs.ResetAll()

		if len(creds["login"]) == 0 || len(creds["pwd"]) == 0 {
			mntl.PrintAttributesAt(11, 1, "Pseudo ou MDP vides", minigo.InversionFond)
			inputs.ActivateFirst()
			return nil, minigo.NoOp
		}

		logged := UsersDb.LogUser(creds["login"], creds["pwd"])
		delete(creds, "pwd")

		if logged {
			logs.InfoLog("sign-in: logged as user=%s\n", creds["login"])
			return creds, minigo.EnvoiOp
		} else {
			mntl.PrintAttributesAt(11, 1, "Pseudo ou MDP invalides", minigo.InversionFond)
			inputs.ActivateFirst()
			return nil, minigo.NoOp
		}
	})

	signInPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	signInPage.SetGuideFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.GuideOp
	})

	signInPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.ActivateNext()
		return nil, minigo.NoOp
	})

	signInPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.ActivatePrev()
		return nil, minigo.NoOp
	})

	return signInPage
}
