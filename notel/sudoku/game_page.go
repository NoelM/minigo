package sudoku

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/NoelM/minigo"
	"github.com/jedib0t/go-sudoku/generator"
	sdk "github.com/jedib0t/go-sudoku/sudoku"
	"github.com/jedib0t/go-sudoku/sudoku/difficulty"
)

func RunPageGame(mntl *minigo.Minitel, login string, level int) (op int) {

	gamePage := minigo.NewPage("sudoku:game", mntl, nil)
	matrix := minigo.NewMatrix(9, 9)
	var grid *sdk.Grid

	start := time.Now()

	gamePage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.ClavierEtendu()

		var d difficulty.Difficulty
		var dName string
		switch level {
		case 1:
			d = difficulty.Easy
			dName = "FACILE"
		case 2:
			d = difficulty.Medium
			dName = "MOYEN"
		case 3:
			d = difficulty.Hard
			dName = "DIFFICILE"
		case 4:
			d = difficulty.Insane
			dName = "EXTREME"
		}
		mntl.WriteStringLeftAt(2, "Niveau:")
		mntl.WriteStringLeftAt(3, dName)

		gen := generator.BackTrackingGenerator(generator.WithRNG(rand.New(rand.NewSource(time.Now().UnixNano()))))

		grid, _ = gen.Generate(nil)
		grid.ApplyDifficulty(d)

		array := grid.MarshalArray()

		lineRef := 2
		colRef := 9
		padding := 2

		// Grid
		mntl.Rect(lineRef, colRef, 9*padding+3, 9*padding+3)
		mntl.VLine(lineRef+1, colRef+3*padding+1, 9*padding+1, minigo.VCenter)
		mntl.VLine(lineRef+1, colRef+6*padding+2, 9*padding+1, minigo.VCenter)
		mntl.HLine(lineRef+3*padding+1, colRef+1, 9*padding+2, minigo.HCenter)
		mntl.HLine(lineRef+6*padding+2, colRef+1, 9*padding+2, minigo.HCenter)

		// Numbers
		mntl.WriteAttributes(minigo.InversionFond)
		for line := range array {
			if line%3 == 0 {
				lineRef += 1
			}
			linePos := lineRef + padding*line

			for col, val := range array[line] {
				if col%3 == 0 {
					colRef += 1
				}
				colPos := colRef + padding*col

				if val == 0 {
					matrix.SetInput(line, col, minigo.NewInput(mntl, linePos, colPos, 1, 1, true))
				} else {
					mntl.WriteStringAt(linePos, colPos, fmt.Sprintf("%d", val))
				}
			}
			colRef -= 3
		}
		mntl.WriteAttributes(minigo.FondNormal)

		mntl.WriteStringLeftAt(24, "Naviguez ←↑→↓")
		mntl.WriteHelperRightAt(24, "Vérif. grille", "ENVOI")
		matrix.InitAll()

		return minigo.NoOp
	})

	gamePage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		mntl.ClavierStandard()
		return nil, minigo.SommaireOp
	})

	gamePage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		for id, s := range matrix.ToArray() {
			if s == "" {
				continue
			}

			val, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				continue
			}

			row := id / 9
			col := id - 9*row
			if !grid.Set(row, col, int(val)) {
				mntl.WriteStringLeftAt(1, fmt.Sprintf("Element invalide: ligne: %d, col: %d", row+1, col+1))
				time.Sleep(2 * time.Second)
				mntl.CleanLine()

				matrix.ActivateFirst()
				return nil, minigo.NoOp
			}
		}

		if !grid.Done() {
			mntl.WriteStringLeftAt(1, "Grille incomplète")
			time.Sleep(2 * time.Second)
			mntl.CleanLine()

			matrix.ActivateFirst()
			return nil, minigo.NoOp

		}

		if ok, err := grid.Validate(); err != nil || !ok {
			mntl.WriteStringLeftAt(1, "Grille invalide")
			time.Sleep(2 * time.Second)
			mntl.CleanLine()

			matrix.ActivateFirst()
			return nil, minigo.NoOp

		} else {
			mntl.WriteStringLeftAt(1, fmt.Sprintf("Bravo %s ! C'est réussi en %s", login, time.Since(start).Round(time.Second).String()))
			time.Sleep(2 * time.Second)
		}

		mntl.ClavierStandard()
		return nil, minigo.EnvoiOp
	})

	gamePage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		matrix.CorrectionActive()
		return nil, minigo.NoOp
	})

	gamePage.SetHautFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		matrix.ActivateUp()
		return nil, minigo.NoOp
	})

	gamePage.SetBasFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		matrix.ActivateDown()
		return nil, minigo.NoOp
	})

	gamePage.SetGaucheFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		matrix.ActivateLeft()
		return nil, minigo.NoOp
	})

	gamePage.SetDroiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		matrix.ActivateRight()
		return nil, minigo.NoOp
	})

	gamePage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key rune) {
		matrix.AppendKeyActive(key)
	})

	_, op = gamePage.Run()
	return op
}
