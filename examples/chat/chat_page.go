package main

import (
	"fmt"

	"github.com/NoelM/minigo"
)

func chatPage(m *minigo.Minitel, envoi chan []byte, messagesList *Messages) {
	userInput := []byte{}

	for {
		select {
		case key := <-m.InKey:
			if key == minigo.Envoi {
				messagesList.AppendTeletelMessage("minitel", userInput)
				envoi <- userInput

				clearInput(m)
				updateScreen(m, messagesList)
				userInput = []byte{}

				m.CursorOnXY(1, InputLine)

			} else if key == minigo.Repetition {
				updateScreen(m, messagesList)
				updateInput(m, userInput)

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
			return
		}
	}
}

const InputLine = 22

func clearInput(m *minigo.Minitel) {
	m.CursorOff()
	m.CleanScreenFromXY(1, InputLine)
}

func updateScreen(m *minigo.Minitel, list *Messages) {
	currentLine := 1

	list.Mtx.RLock()
	defer list.Mtx.RUnlock()

	m.CursorOff()
	for i := len(list.List) - 1; i >= 0; i -= 1 {
		// 3 because the format is: "nick > text"
		msgLen := len(list.List[i].Nick) + len(list.List[i].Text) + 3

		// 2 because if msgLen < 40, the division gives 0 and one breaks another line for readability
		// nick > text
		// <blank>
		// nick > text2
		msgLines := msgLen/40 + 2

		if currentLine+msgLines > InputLine {
			break
		}

		buf := minigo.GetMoveCursorXY(0, currentLine)
		buf = append(buf, minigo.EncodeSprintf("%s > ", list.List[i].Nick)...)

		if list.List[i].Type == Message_Teletel {
			buf = append(buf, list.List[i].Text...)
		} else {
			buf = append(buf, minigo.EncodeMessage(list.List[i].Text)...)
		}

		buf = append(buf, minigo.GetCleanLineFromCursor()...)
		buf = append(buf, minigo.GetMoveCursorReturn(1)...)
		buf = append(buf, minigo.GetCleanLine()...)
		m.Send(buf)

		currentLine += msgLines
	}
}

func appendInput(m *minigo.Minitel, inputLen int, key byte) {
	y := inputLen / 40
	x := inputLen % 40

	buf := minigo.GetMoveCursorXY(x+1, y+InputLine)
	buf = append(buf, key)
	m.Send(buf)
}

func corrInput(m *minigo.Minitel, inputLen int) {
	y := (inputLen - 1) / 40
	x := (inputLen - 1) % 40

	buf := minigo.GetMoveCursorXY(x+1, y+InputLine)
	buf = append(buf, minigo.GetCleanLineFromCursor()...)
	m.Send(buf)
}

func updateInput(m *minigo.Minitel, userInput []byte) {
	m.WriteBytesXY(1, InputLine, userInput)
}
