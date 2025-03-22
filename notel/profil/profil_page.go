package profil

import (
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func RunPageProfil(mntl *minigo.Minitel, userDB *databases.UsersDatabase, pseudo string) (op int) {
	profilPage := minigo.NewPage("profil", mntl, nil)

	profilPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.Reset()

		usr, err := userDB.LoadUser(pseudo)
		if err != nil {
			mntl.Print("Impossible de charger le profil")
			time.Sleep(2 * time.Second)
			return minigo.SommaireOp
		}

		mntl.MoveAt(2, 0)
		mntl.PrintAttributes("Profil", minigo.DoubleHauteur)

		mntl.Return(1) // Row 3
		mntl.HLine(40, minigo.HCenter)

		mntl.Return(1) // Row 5
		mntl.Attributes(minigo.FondVert, minigo.CaractereNoir)
		mntl.Print(" PSEUDO")
		mntl.Attributes(minigo.FondNoir, minigo.CaractereBlanc)
		mntl.Print(" " + pseudo)

		mntl.Return(1) // Row 6
		mntl.Attributes(minigo.FondVert, minigo.CaractereNoir)
		mntl.Print(" BIO")
		mntl.Attributes(minigo.FondNoir, minigo.CaractereBlanc)
		mntl.Print(" ")
		inputs.AppendInput("bio", minigo.NewInputWithValue(mntl, usr.Bio, 7, 0, 39, 2, true))

		mntl.Return(4) // Row 10
		mntl.Attributes(minigo.FondVert, minigo.CaractereNoir)
		mntl.Print(" SERVEUR MINITEL")
		mntl.Attributes(minigo.FondNoir, minigo.CaractereBlanc)
		mntl.Print(" ")
		inputs.AppendInput("tel", minigo.NewInputWithValue(mntl, usr.Tel, 11, 0, 39, 1, true))

		mntl.Return(3) // Row 13
		mntl.Attributes(minigo.FondVert, minigo.CaractereNoir)
		mntl.Print(" LIEU")
		mntl.Attributes(minigo.FondNoir, minigo.CaractereBlanc)
		mntl.Print(" ")
		inputs.AppendInput("loc", minigo.NewInputWithValue(mntl, usr.Location, 14, 0, 39, 1, true))

		mntl.Return(3) // Row 16
		mntl.Attributes(minigo.FondVert, minigo.CaractereNoir)
		mntl.Print(" AFFICHAGE REPERTOIRE (OUI/NON)")
		mntl.Attributes(minigo.FondNoir, minigo.CaractereBlanc)
		mntl.Print(" ")
		appAnnuString := "NON"
		if usr.AppAnnuaire {
			appAnnuString = "OUI"
		}
		inputs.AppendInput("annu", minigo.NewInputWithValue(mntl, appAnnuString, 16, 32, 3, 1, true))

		mntl.Return(2)
		mntl.Helper("Nav.", "RETOUR/SUITE", minigo.FondBleu, minigo.CaractereBlanc)
		mntl.Right(5)
		mntl.HelperRight("Sauver", "ENVOI", minigo.FondVert, minigo.CaractereNoir)

		mntl.Return(1)
		mntl.HLine(40, minigo.HCenter)

		list := minigo.NewList(mntl, 21, 1, 25, 1)
		list.AppendItem(mdpKey, "Changer mot de passe")
		list.AppendItem(supprKey, "Supprimer le compte")
		list.Display()

		mntl.MoveAt(23, 1)
		mntl.Helper("CODE .... +", "ENVOI", minigo.FondVert, minigo.CaractereNoir)
		inputs.AppendInput("code", minigo.NewInput(mntl, 23, 6, 4, 1, true))

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
		data := inputs.ToMap()
		if code, ok := data["code"]; ok {
			switch code {
			case mdpKey:
				return nil, mdpId
			case supprKey:
				return nil, supprId
			}
		}

		if data["annu"] != "OUI" && data["annu"] != "NON" {
			mntl.MoveAt(4, 0)
			mntl.PrintAttributes("Champ affichage annuaire faux", minigo.InversionFond)

			time.Sleep(2 * time.Second)
			return nil, minigo.SommaireOp
		}

		usr, err := userDB.LoadUser(pseudo)
		if err != nil {
			mntl.MoveAt(4, 0)
			mntl.Print("Impossible de charger le profil")

			time.Sleep(2 * time.Second)
			return nil, minigo.SommaireOp
		}

		usr.Bio = data["bio"]
		usr.Tel = data["tel"]
		usr.Location = data["loc"]
		usr.AppAnnuaire = data["annu"] == "OUI"

		err = userDB.SetUser(usr)
		if err != nil {
			mntl.MoveAt(4, 0)
			mntl.Print("Impossible d'enregistrer le profil")

			time.Sleep(2 * time.Second)
			return nil, minigo.SommaireOp
		}

		mntl.MoveAt(4, 0)
		mntl.Print("Modifications effectuées avec succès !")

		time.Sleep(2 * time.Second)
		return nil, minigo.SommaireOp
	})

	_, op = profilPage.Run()
	return op
}
