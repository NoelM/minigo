package main

import (
	"fmt"

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
		case key := <-m.InKey:
			if key == minigo.Envoi {
				if len(nickInput.Value) == 0 {
					continue
				}
				m.Reset()
				return nickInput.Value

			} else if key == minigo.Correction {
				nickInput.Correction()

			} else if minigo.IsUintAValidChar(key) {
				nickInput.AppendKey(byte(key))

			} else {
				fmt.Printf("key: %d not supported", key)
			}
		default:
			continue
		}

		if m.ContextError() != nil {
			return nil
		}
	}
}
