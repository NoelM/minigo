package main

import (
	"fmt"
	"time"

	"github.com/NoelM/minigo"
)

func chatPage(m *minigo.Minitel, ircDvr *IrcDriver) int {
	infoLog.Printf("Opening chat page for nick=%s\n", ircDvr.Nick)

	messages := []Message{}
	messageInput := minigo.NewInput(m, 1, InputLine, 40, 5, ">", true)

	m.WriteStringXY(1, 1, fmt.Sprintf(">>> CONNECTE '%s' SUR #MINITEL", ircDvr.Nick))
	time.Sleep(2 * time.Second)
	m.CleanLine()

	messageInput.Repetition()
	m.RouleauOn()

	lastId := 0
	for {
		select {
		case key := <-m.RecvKey:
			if key == minigo.Envoi {
				msg := Message{
					Nick: ircDvr.Nick,
					Text: string(messageInput.Value),
					Type: Message_Teletel,
					Time: time.Now(),
				}
				messages = append(messages, msg)
				ircDvr.SendMessage <- msg

				infoLog.Printf("Send new message to IRC from nick=%s len=%d\n", ircDvr.Nick, len(msg.Text))

				messageInput.Clear()
				updateScreen(m, messages, &lastId)

				messageInput.Repetition()

			} else if key == minigo.Repetition {
				infoLog.Printf("User nick=%s asked for a refresh\n", ircDvr.Nick)

				messageInput.ClearScreen()
				updateScreen(m, messages, &lastId)
				messageInput.Repetition()

			} else if key == minigo.Correction {
				messageInput.Correction()

			} else if key == minigo.Sommaire {
				return sommaireId

			} else if minigo.IsUintAValidChar(key) {
				messageInput.AppendKey(byte(key))

			} else {
				errorLog.Printf("Not supported key: %d\n", key)
			}

		case msg := <-ircDvr.RecvMessage:
			messages = append(messages, msg)

		case <-m.Quit:
			warnLog.Printf("Quitting chatpage for nick: %s\n", ircDvr.Nick)
			return quitId

		default:
			continue
		}
	}
}

const InputLine = 22

func updateScreen(m *minigo.Minitel, list []Message, lastId *int) {
	m.CursorOff()
	for i := *lastId; i < len(list); i += 1 {
		// 3 because the format is: "nick > text"
		msgLen := len(list[i].Nick) + len(list[i].Text) + 3

		// 2 because if msgLen < 40, the division gives 0 and one breaks another line for readability
		// nick > text
		// <blank>
		// nick > text2
		msgLines := msgLen/40 + 2

		buf := minigo.GetMoveCursorXY(1, 24)
		for k := 0; k < msgLines; k += 1 {
			buf = append(buf, minigo.GetMoveCursorReturn(1)...)
		}
		buf = append(buf, minigo.GetMoveCursorXY(1, InputLine-msgLines)...)
		buf = append(buf, minigo.EncodeSprintf("%s > ", list[i].Nick)...)

		if list[i].Type == Message_Teletel {
			buf = append(buf, list[i].Text...)
		} else {
			buf = append(buf, minigo.EncodeMessage(list[i].Text)...)
		}
		m.Send(buf)
	}

	*lastId = len(list)

	m.WriteStringXY(1, 24, "MAJ ECRAN ")
	m.WriteAttributes(minigo.InversionFond)
	m.Send(minigo.EncodeMessage("REPET."))
	m.WriteAttributes(minigo.FondNormal)

	m.WriteStringXY(25, 24, "MESSAGE + ")
	m.WriteAttributes(minigo.InversionFond)
	m.Send(minigo.EncodeMessage("ENVOI"))
	m.WriteAttributes(minigo.FondNormal)
}
