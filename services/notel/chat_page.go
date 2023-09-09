package main

import (
	"fmt"
	"time"

	"github.com/NoelM/minigo"
)

func NewChatPage(m *minigo.Minitel, ircDrv *IrcDriver) *minigo.Page {
	chatPage := minigo.NewPage("chat", m, nil)

	messages := []Message{}
	lastMsgId := 0

	chatPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) {
		infoLog.Printf("opening chat page for nick=%s\n", ircDrv.Nick)

		inputs.AppendInput("messages", minigo.NewInput(m, 1, InputLine, 40, 5, ">", true))

		m.WriteStringXY(1, 1, fmt.Sprintf(">>> CONNECTE '%s' SUR #MINITEL", ircDrv.Nick))
		time.Sleep(2 * time.Second)
		m.CleanLine()

		helpers(m)

		inputs.RepetitionActive()
		m.RouleauOn()
	})

	chatPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		msg := Message{
			Nick: ircDrv.Nick,
			Text: string(inputs.ValueActive()),
			Type: Message_Teletel,
			Time: time.Now(),
		}
		messages = append(messages, msg)
		ircDrv.SendMessage <- msg

		infoLog.Printf("send new message to IRC from nick=%s len=%d\n", ircDrv.Nick, len(msg.Text))

		inputs.ClearActive()
		updateScreen(m, messages, &lastMsgId)

		inputs.RepetitionActive()

		return nil, minigo.NoOp
	})

	chatPage.SetRepetitionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		infoLog.Printf("user nick=%s asked for a refresh\n", ircDrv.Nick)

		inputs.ClearScreenAll()
		updateScreen(m, messages, &lastMsgId)
		inputs.RepetitionAll()

		return nil, minigo.NoOp
	})

	chatPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	chatPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	return chatPage
}

const InputLine = 22

func updateScreen(m *minigo.Minitel, list []Message, lastId *int) {
	m.CursorOff()
	for i := *lastId; i < len(list); i += 1 {
		msgLen := 5 + len(list[i].Nick) + 4 + len(list[i].Text)

		// 1 because if msgLen < 40, the division gives 0 and one breaks another line for readability
		// nick > text
		// nick > text2
		msgLines := msgLen/40 + 1

		buf := minigo.GetMoveCursorXY(1, 24)
		buf = append(buf, minigo.GetMoveCursorReturn(msgLines)...)
		buf = append(buf, minigo.GetMoveCursorXY(1, InputLine-msgLines)...)

		buf = append(buf, minigo.EncodeAttributes(minigo.InversionFond)...)
		buf = append(buf, minigo.EncodeSprintf("%s | %s ", list[i].Time.Format("15:04"), list[i].Nick)...)
		buf = append(buf, minigo.EncodeAttributes(minigo.FondNormal)...)

		if list[i].Type == Message_Teletel {
			buf = append(buf, list[i].Text...)
		} else {
			buf = append(buf, minigo.EncodeMessage(list[i].Text)...)
		}
		m.Send(buf)
	}

	*lastId = len(list)

	helpers(m)
}

func helpers(m *minigo.Minitel) {
	m.WriteStringXY(1, 24, "MAJ ECRAN ")
	m.WriteAttributes(minigo.InversionFond)
	m.Send(minigo.EncodeMessage("REPET."))
	m.WriteAttributes(minigo.FondNormal)

	m.WriteStringXY(25, 24, "MESSAGE + ")
	m.WriteAttributes(minigo.InversionFond)
	m.Send(minigo.EncodeMessage("ENVOI"))
	m.WriteAttributes(minigo.FondNormal)
}
