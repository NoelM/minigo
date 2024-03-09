package meteo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/NoelM/minigo/notel/logs"
)

func NewCommunesPage(mntl *minigo.Minitel, cDB *databases.CommuneDatabase, selectedCP map[string]string) *minigo.Page {
	communesPage := minigo.NewPage("communes", mntl, selectedCP)

	var communes []databases.Commune

	communesPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		codePostal, ok := initData["code_postal"]
		if !ok {
			return minigo.SommaireOp
		}

		communes = cDB.GetCommunesFromCodePostal(codePostal)
		if communes == nil {
			mntl.WriteString("Impossible de trouver une commune")
			mntl.Return(1)
			mntl.WriteString("Retour au sommaire dans 5 sec.")

			time.Sleep(5 * time.Second)
			return minigo.SommaireOp
		}

		mntl.MoveAt(2, 0)
		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteString("Choisissez une commune:")
		mntl.WriteAttributes(minigo.GrandeurNormale)

		var listItem []string
		for _, c := range communes {
			listItem = append(listItem, fmt.Sprintf("%s (%02s)", c.NomCommune, c.CodeDepartement))
		}
		communeList := minigo.NewListEnum(mntl, listItem, 4, 1, 20, 1)

		communeList.Display()

		mntl.MoveAt(len(communes)+5, 1)
		mntl.PrintHelper("CHOIX: .. +", "ENVOI", minigo.FondVert, minigo.CaractereNoir)
		inputs.AppendInput("commune_id", minigo.NewInput(mntl, len(communes)+5, 8, 2, 1, true))

		mntl.MoveAt(24, 0)
		mntl.PrintHelperRight("Retour choix Code Postal →", "SOMMAIRE", minigo.FondCyan, minigo.CaractereNoir)

		inputs.ActivateFirst()

		return minigo.NoOp
	})

	communesPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if len(inputs.ValueActive()) == 0 {
			logs.WarnLog("empty commune choice\n")
			return nil, minigo.NoOp
		}

		communeId, err := strconv.ParseInt(inputs.ValueActive(), 10, 32)
		if err != nil {
			logs.ErrorLog("unable to parse code choice: %s\n", err.Error())
			return nil, minigo.SommaireOp
		}

		if communeId > 0 && int(communeId-1) >= len(communes) {
			logs.ErrorLog("choice %d out of range\n", communeId)
			return nil, minigo.SommaireOp
		}
		logs.InfoLog("chosen commune: %s\n", communes[communeId-1].NomCommune)

		data, err := json.Marshal(communes[communeId-1])
		if err != nil {
			logs.ErrorLog("unable to marshall JSON: %s\n", err.Error())
			return nil, minigo.SommaireOp
		}

		return map[string]string{"commune": string(data)}, minigo.QuitOp
	})

	communesPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key rune) {
		inputs.AppendKeyActive(key)
	})

	communesPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	communesPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	return communesPage
}
