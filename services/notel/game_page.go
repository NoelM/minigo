package main

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

func NewPageGame(mntl *minigo.Minitel, level map[string]string) *minigo.Page {

	gamePage := minigo.NewPage("sudoku:game", mntl, level)
	matrix := minigo.NewMatrix(9, 9)
	var grid *sdk.Grid

	gamePage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		level, err := strconv.ParseInt(initData["level"], 10, 32)
		if err != nil {
			return minigo.QuitOp
		}

		var d difficulty.Difficulty
		switch level {
		case 1:
			d = difficulty.Easy
		case 2:
			d = difficulty.Medium
		case 3:
			d = difficulty.Hard
		case 4:
			d = difficulty.Insane
		}

		var rnd *rand.Rand
		gen := generator.BackTrackingGenerator(generator.WithRNG(rnd))

		grid, _ = gen.Generate(nil)
		grid.ApplyDifficulty(d)

		array := grid.MarshalArray()

		lineRef := 5
		colRef := 2
		padding := 2

		for line := range array {
			linePos := lineRef + padding*line

			for col, val := range array[line] {
				colPos := colRef + padding*col

				if val == 0 {
					infoLog.Printf("input at %d %d\n", linePos, colPos)
					matrix.SetInput(line, col, minigo.NewInput(mntl, linePos, colPos, 1, 1, true))
				} else {
					infoLog.Printf("value=%d at %d %d\n", val, linePos, colPos)
					mntl.WriteStringAt(linePos, colPos, fmt.Sprintf("%d", val))
				}
			}
		}

		matrix.InitAll()
		matrix.ActivateFirst()

		mntl.WriteStringLeft(24, "NAVIGUEZ ←↑→↓")
		mntl.WriteHelperRight(24, "VALIDEZ", "ENVOI")

		return minigo.NoOp
	})

	gamePage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
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
			grid.Set(row, col, int(val))
		}

		if ok, err := grid.Validate(); err != nil || !ok {
			mntl.WriteStringLeft(2, "Grille invalide...")
			return nil, minigo.NoOp
		} else {
			mntl.WriteStringLeft(2, "BRAVO ! Réussi")
			time.Sleep(2 * time.Second)
		}

		return nil, minigo.EnvoiOp
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

	return gamePage
}
