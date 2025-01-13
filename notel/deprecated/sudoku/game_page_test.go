package sudoku

import (
	"math/rand"
	"testing"

	"github.com/jedib0t/go-sudoku/generator"
	"github.com/jedib0t/go-sudoku/sudoku/difficulty"
)

func TestSudokuFill(t *testing.T) {
	gen := generator.BackTrackingGenerator(generator.WithRNG(rand.New(rand.NewSource(1))))

	grid, _ := gen.Generate(nil)
	grid.ApplyDifficulty(difficulty.Easy)
	/*
		6,0,3,2,0,0,9,0,1
		9,0,8,3,6,4,2,0,7
		4,0,0,1,7,9,3,8,0
		0,4,0,7,0,0,0,0,0
		0,0,0,8,5,0,7,1,4
		0,0,0,4,9,1,0,3,5
		7,9,4,5,0,8,1,6,0
		0,0,2,0,0,7,5,0,8
		8,5,0,9,0,3,0,0,0
	*/

	if grid.Done() {
		t.Fatal("grid not done")
	}

	if !grid.Set(0, 1, 7) {
		t.Fatal("cannot add value")
	}
}
