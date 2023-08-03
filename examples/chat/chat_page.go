package main

import (
	"fmt"
	"time"

	"github.com/NoelM/minigo"
)

func chatPage(m *minigo.Minitel, nick string, envoi chan []byte, messagesList *Messages) {
	messageInput := minigo.NewInput(m, 1, InputLine, 40, 5, ">", true)

	m.WriteStringXY(1, 1, fmt.Sprintf(">>> CONNECTE '%s' SUR #MINITEL", nick))
	time.Sleep(2 * time.Second)
	m.CleanLine()

	messageInput.Repetition()
	m.RouleauOn()

	lastId := 0
	for {
		select {
		case key := <-m.InKey:
			if key == minigo.Envoi {
				messagesList.AppendTeletelMessage("minitel", messageInput.Value)
				envoi <- messageInput.Value

				messageInput.Clear()
				updateScreen(m, messagesList, &lastId)

				messageInput.Repetition()

			} else if key == minigo.Repetition {
				messageInput.ClearScreen()
				updateScreen(m, messagesList, &lastId)
				messageInput.Repetition()

			} else if key == minigo.Correction {
				messageInput.Correction()

			} else if minigo.IsUintAValidChar(key) {
				messageInput.AppendKey(byte(key))

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

func updateScreen(m *minigo.Minitel, list *Messages, lastId *int) {
	list.Mtx.RLock()
	defer list.Mtx.RUnlock()

	m.CursorOff()
	for i := *lastId - 1; i < len(list.List); i += 1 {
		// 3 because the format is: "nick > text"
		msgLen := len(list.List[i].Nick) + len(list.List[i].Text) + 3

		// 2 because if msgLen < 40, the division gives 0 and one breaks another line for readability
		// nick > text
		// <blank>
		// nick > text2
		msgLines := msgLen/40 + 2

		buf := minigo.GetMoveCursorXY(1, 24)
		buf = append(buf, minigo.GetMoveCursorReturn(msgLines)...)
		buf = append(buf, minigo.GetMoveCursorXY(1, InputLine-msgLines)...)
		buf = append(buf, minigo.EncodeSprintf("%s > ", list.List[i].Nick)...)

		if list.List[i].Type == Message_Teletel {
			buf = append(buf, list.List[i].Text...)
		} else {
			buf = append(buf, minigo.EncodeMessage(list.List[i].Text)...)
		}
		m.Send(buf)
	}

	*lastId = len(list.List)
}
