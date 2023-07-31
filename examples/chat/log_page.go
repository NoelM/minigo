package main

import (
	"fmt"

	"github.com/NoelM/minigo"
)

func logPage(m *minigo.Minitel) []byte {
	userInput := []byte{}

	m.WriteAttributes(minigo.DoubleGrandeur, minigo.InversionFond)
	m.WriteStringXY(10, 10, "MINI-CHAT")

	m.WriteAttributes(minigo.GrandeurNormale, minigo.FondNormal)
	m.WriteStringXY(10, 12, "PSEUDO : ")
	m.Return(1)
	m.CursorOn()

	for {
		select {
		case key := <-m.InKey:
			if key == minigo.Envoi {
				if len(userInput) == 0 {
					continue
				}
				return userInput

			} else if key == minigo.Correction {
				if len(userInput) > 0 {
					corrInput(m, len(userInput))
					userInput = userInput[:len(userInput)-1]
				}

			} else if minigo.IsUintAValidChar(key) {
				appendInput(m, len(userInput), byte(key))
				userInput = append(userInput, byte(key))

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
