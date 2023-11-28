package main

import (
	"strconv"

	"github.com/NoelM/minigo"
)

func NewPageLevel(mntl *minigo.Minitel) *minigo.Page {
	levelPage := minigo.NewPage("sudoku:level", mntl, nil)

	levelPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringLeft(2, "SUDOKU")
		mntl.WriteAttributes(minigo.GrandeurNormale)

		list := minigo.NewListEnum(mntl, []string{"Facile", "Moyen", "Difficile", "Extrême"})
		list.SetXY(1, 5)
		list.SetEntryHeight(1)
		list.Display()

		inputs.AppendInput("level", minigo.NewInput(mntl, 12, 8, 1, 1, true))

		mntl.WriteHelperLeft(12, "NIVEAU . +", "ENVOI")
		inputs.ActivateFirst()

		return minigo.NoOp
	})

	levelPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	levelPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		var err error
		var level int64
		if v := inputs.ValueActive(); v != "" {
			if level, err = strconv.ParseInt(v, 10, 32); err != nil {
				inputs.ResetAll()
				return nil, minigo.NoOp
			}
		}

		if level < 0 && level > 4 {
			inputs.ResetAll()
			return nil, minigo.NoOp
		}

		return inputs.ToMap(), minigo.EnvoiOp
	})

	levelPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return map[string]string{"level": "2"}, minigo.EnvoiOp
	})

	levelPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key rune) {
		inputs.AppendKeyActive(key)
	})

	return levelPage
}
