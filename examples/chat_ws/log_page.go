package main

import (
	"github.com/NoelM/minigo"
)

func logPage(m *minigo.Minitel) []byte {
	nickInput := minigo.NewInput(m, 10, 13, 10, 1, "", true)

	m.WriteAttributes(minigo.DoubleGrandeur, minigo.InversionFond)
	m.WriteStringXY(10, 10, "MINI-CHAT")

	m.WriteAttributes(minigo.GrandeurNormale, minigo.FondNormal)
	m.WriteStringXY(10, 12, "PSEUDO : ")
	m.CursorOnXY(10, 13)

	for {
		select {
		case key := <-m.RecvKey:
			if key == minigo.Envoi {
				if len(nickInput.Value) == 0 {
					warnLog.Println("Empty nick input")
					continue
				}
				m.Reset()

				infoLog.Printf("Logged as: %s\n", nickInput.Value)
				return nickInput.Value

			} else if key == minigo.Correction {
				nickInput.Correction()

			} else if minigo.IsUintAValidChar(key) {
				nickInput.AppendKey(byte(key))

			} else {
				errorLog.Printf("Not supported key: %d\n", key)
			}

		case <-m.Quit:
			warnLog.Println("Quitting log page")
			return nil

		default:
			continue
		}
	}
}
