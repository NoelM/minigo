package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/logs"
)

func NewCommunesPage(mntl *minigo.Minitel, selectedCP map[string]string) *minigo.Page {
	communesPage := minigo.NewPage("communes", mntl, selectedCP)

	var communes []Commune

	communesPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		codePostal, ok := initData["code_postal"]
		if !ok {
			return sommaireId
		}

		communes = CommuneDb.GetCommunesFromCodePostal(codePostal)
		if communes == nil {
			mntl.CleanScreen()
			mntl.WriteStringLeft(1, "IMPOSSIBLE DE TROUVER UNE COMMUNE")
			mntl.WriteStringLeft(2, "RETOUR AU SOMMAIRE DANS 5 SEC.")
			time.Sleep(5 * time.Second)
			return sommaireId
		}

		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringLeft(2, "CHOISISSEZ UNE COMMUNE:")
		mntl.WriteAttributes(minigo.GrandeurNormale)

		communeList := minigo.NewListEnum(mntl, nil)
		communeList.SetXY(1, 4)
		communeList.SetEntryHeight(1)

		for _, c := range communes {
			communeList.AppendItem(fmt.Sprintf("%s (%02s)", c.NomCommune, c.CodeDepartement))
		}
		communeList.Display()

		mntl.WriteHelperLeft(len(communes)+5, "CHOIX: .. +", "ENVOI")

		mntl.WriteHelperLeft(24, "CHOIX CODE POSTAL", "SOMMAIRE")

		inputs.AppendInput("commune_id", minigo.NewInput(mntl, len(communes)+5, 8, 2, 1, true))
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
			return nil, sommaireId
		}

		if communeId > 0 && int(communeId-1) >= len(communes) {
			logs.ErrorLog("choice %d out of range\n", communeId)
			return nil, sommaireId
		}
		logs.InfoLog("chosen commune: %s\n", communes[communeId-1].NomCommune)

		data, err := json.Marshal(communes[communeId-1])
		if err != nil {
			logs.ErrorLog("unable to marshall JSON: %s\n", err.Error())
			return nil, sommaireId
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
		return nil, sommaireId
	})

	return communesPage
}
