package main

import (
	"time"

	"github.com/NoelM/minigo"
)

func StartSplash(m *minigo.Minitel) {
	m.CleanScreen()
	m.MoveAt(1, 0)

	m.Print("→ Bienvenue sur NOTEL")
	m.Return(1)
	m.Print("Mesure de la stabilité de la ligne:")
	m.Return(1)

	for i := 0; i < 5; i += 1 {
		time.Sleep(500 * time.Millisecond)
		m.Printf("%d", i+1)
		m.Right(1)
	}
}
