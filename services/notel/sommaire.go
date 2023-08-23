package main

import (
	"strconv"

	"github.com/NoelM/minigo"
)

const (
	sommaire = iota
	irc
	meteo
	guide
)

func PageSommaire(m *minigo.Minitel) int {
	entry := minigo.NewInput(m, 32, 24, 2, 1, "", true)

	m.CleanScreen()
	m.SendVDT("static/notel.vdt")

	m.WriteAttributes(minigo.GrandeurNormale, minigo.InversionFond)
	m.WriteStringXY(1, 8, " 1 ")
	m.WriteAttributes(minigo.FondNormal)
	m.WriteStringXY(5, 8, "Mini-Chat (IRC)")

	m.CursorOnXY(32, 24)

	for {
		select {
		case key := <-m.RecvKey:
			if key == minigo.Envoi {
				if len(entry.Value) == 0 {
					warnLog.Println("empty choice")
					continue
				}
				m.Reset()

				infoLog.Printf("choose service: %s\n", entry.Value)
				id, err := strconv.Atoi(string(entry.Value))
				if err != nil {
					warnLog.Println("unable to parse choice")
					return 0
				}
				return id

			} else if key == minigo.Correction {
				entry.Correction()

			} else if minigo.IsUintAValidChar(key) {
				entry.AppendKey(byte(key))

			} else {
				errorLog.Printf("Not supported key: %d\n", key)
			}

		case <-m.Quit:
			warnLog.Println("Quitting log page")
			return -1

		default:
			continue
		}
	}
}
