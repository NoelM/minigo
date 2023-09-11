package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/NoelM/minigo"
)

func NewCommunesPage(mntl *minigo.Minitel, codePostal map[string]string) *minigo.Page {
	communesPage := minigo.NewPage("communes", mntl, codePostal)

	var communes []Commune

	communesPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		codePostal, ok := initData["code_postal"]
		if !ok {
			return sommaireId
		}

		communes = getCommunesFromCodePostal(codePostal)
		if communes == nil {
			mntl.CleanScreen()
			mntl.WriteStringAt(1, 1, "IMPOSSIBLE DE TROUVER UNE COMMUNE")
			mntl.WriteStringAt(1, 2, "RETOUR AU SOMMAIRE DANS 5 SEC.")
			time.Sleep(5 * time.Second)
			return sommaireId
		}

		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringAt(1, 2, "CHOISISSEZ UNE COMMUNE:")
		mntl.WriteAttributes(minigo.GrandeurNormale)

		communeList := minigo.NewList(mntl, nil)
		communeList.SetXY(1, 4)
		communeList.SetEntryHeight(1)

		for _, c := range communes {
			communeList.AppendItem(fmt.Sprintf("%s (%02s)", c.NomCommune, c.CodeDepartement))
		}
		communeList.Display()

		mntl.WriteStringAt(1, len(communes)+5, "CHOIX: .. + ")
		mntl.WriteAttributes(minigo.InversionFond)
		mntl.Send(minigo.EncodeMessage("ENVOI"))
		mntl.WriteAttributes(minigo.FondNormal)

		mntl.WriteStringAt(1, 24, "CHOIX CODE POSTAL ")
		mntl.WriteAttributes(minigo.InversionFond)
		mntl.Send(minigo.EncodeMessage("SOMMAIRE"))
		mntl.WriteAttributes(minigo.FondNormal)

		inputs.AppendInput("commune_id", minigo.NewInput(mntl, 8, len(communes)+5, 2, 1, "", true))
		inputs.ActivateFirst()

		return minigo.NoOp
	})

	communesPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if len(inputs.ValueActive()) == 0 {
			warnLog.Println("empty commune choice")
			return nil, minigo.NoOp
		}

		communeId, err := strconv.ParseInt(string(inputs.ValueActive()), 10, 32)
		if err != nil {
			errorLog.Printf("unable to parse code choice: %s\n", err.Error())
			return nil, sommaireId
		}

		if communeId > 0 && int(communeId-1) >= len(communes) {
			errorLog.Printf("choice %d out of range\n", communeId)
			return nil, sommaireId
		}
		infoLog.Printf("chosen commune: %s\n", communes[communeId-1].NomCommune)

		data, err := json.Marshal(communes[communeId-1])
		if err != nil {
			errorLog.Printf("unable to marshall JSON: %s\n", err.Error())
			return nil, sommaireId
		}

		return map[string]string{"commune": string(data)}, minigo.QuitOp
	})

	communesPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key uint) {
		inputs.AppendKeyActive(byte(key))
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
