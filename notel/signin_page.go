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

		mntl.MoveAt(10, 1)
		mntl.WriteAttributes(minigo.FondNormal, minigo.DoubleHauteur)
		mntl.WriteString("Connectez vous au serveur")

		mntl.Return(3)
		mntl.MoveRight(1)
		mntl.WriteString("PSEUDO:")
		inputs.AppendInput("login", minigo.NewInput(mntl, 13, 15, 10, 1, true))

		mntl.Return(1)
		mntl.MoveRight(1)
		mntl.WriteString("MOT DE PASSE:")
		inputs.AppendInput("pwd", minigo.NewInput(mntl, 14, 15, 10, 1, true))

		mntl.Return(2)
		mntl.MoveRight(1)
		mntl.PrintHelperLeftAt(16, "Validez avec", "ENVOI")
		mntl.PrintHelperLeftAt(20, "Nouveau ici ? Appuyez sur", "GUIDE")

		inputs.InitAll()
		return minigo.NoOp
	})

	signInPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key int32) {
		inputs.AppendKeyActive(key)
	})

	signInPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		creds := inputs.ToMap()
		inputs.ResetAll()

		if len(creds["login"]) == 0 || len(creds["pwd"]) == 0 {
			mntl.WriteStringAtWithAttributes(11, 1, "Pseudo ou MDP vides", minigo.InversionFond)
			inputs.ActivateFirst()
			return nil, minigo.NoOp
		}

		logged := UsersDb.LogUser(creds["login"], creds["pwd"])
		delete(creds, "pwd")

		if logged {
			logs.InfoLog("sign-in: logged as user=%s\n", creds["login"])
			return creds, minigo.EnvoiOp
		} else {
			mntl.WriteStringAtWithAttributes(11, 1, "Pseudo ou MDP invalides", minigo.InversionFond)
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
