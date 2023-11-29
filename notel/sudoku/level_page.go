package sudoku

import (
	"strconv"

	"github.com/NoelM/minigo"
)

func RunPageLevel(mntl *minigo.Minitel) (level, op int) {
	levelPage := minigo.NewPage("sudoku:level", mntl, nil)

	levelPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		mntl.WriteAttributes(minigo.DoubleHauteur)
		mntl.WriteStringLeft(2, "SUDOKU")
		mntl.WriteAttributes(minigo.GrandeurNormale)
		mntl.WriteStringLeft(3, "Expérimental, donc des bugs paaaartout")

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
		var parseLevel int64
		if v := inputs.ValueActive(); v != "" {
			if parseLevel, err = strconv.ParseInt(v, 10, 32); err != nil {
				inputs.ResetAll()
				return nil, minigo.NoOp
			}
		}

		if parseLevel < 0 && parseLevel > 4 {
			inputs.ResetAll()
			return nil, minigo.NoOp
		}

		level = int(parseLevel)
		return nil, minigo.EnvoiOp
	})

	levelPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return map[string]string{"level": "2"}, minigo.EnvoiOp
	})

	levelPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key rune) {
		inputs.AppendKeyActive(key)
	})

	_, op = levelPage.Run()
	return level, op
}
