package main

import (
	"time"

	"github.com/NoelM/minigo"
)

func NewChatPage(m *minigo.Minitel, ircDrv *IrcDriver) *minigo.Page {
	chatPage := minigo.NewPage("chat", m, nil)

	subscriberId := 0

	chatPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		infoLog.Printf("opening chat page for nick=%s\n", ircDrv.Nick)

		subscriberId = MessageDb.Subscribe()
		inputs.AppendInput("messages", minigo.NewInput(m, 1, InputLine, 40, 5, ">", true))

		m.RouleauOn()
		updateScreen(m, subscriberId)
		helpers(m)

		inputs.RepetitionActive()

		return minigo.NoOp
	})

	chatPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		msg := Message{
			Nick: ircDrv.Nick,
			Text: string(inputs.ValueActive()),
			Type: MessageTeletel,
			Time: time.Now(),
		}
		MessageDb.PushMessage(msg)
		ircDrv.SendMessage <- msg

		infoLog.Printf("send new message to IRC from nick=%s len=%d\n", ircDrv.Nick, len(msg.Text))

		inputs.ClearActive()
		updateScreen(m, subscriberId)

		inputs.RepetitionActive()

		return nil, minigo.NoOp
	})

	chatPage.SetRepetitionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		infoLog.Printf("user nick=%s asked for a refresh\n", ircDrv.Nick)

		inputs.ClearScreenAll()
		updateScreen(m, subscriberId)
		inputs.RepetitionAll()

		return nil, minigo.NoOp
	})

	chatPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	chatPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		MessageDb.Resign(subscriberId)
		m.RouleauOff()
		return nil, sommaireId
	})

	chatPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key uint) {
		inputs.AppendKeyActive(byte(key))
	})

	return chatPage
}

const InputLine = 22

func updateScreen(m *minigo.Minitel, subscriberId int) {
	m.CursorOff()
	lastMsg := MessageDb.GetMessages(subscriberId)
	firstMsg := len(lastMsg) - InputLine
	if firstMsg < 0 {
		firstMsg = 0
	}
	lastMsg = lastMsg[firstMsg:]

	for _, msg := range lastMsg {
		// 5 because of the date format "15:04"
		// 3 because of " - "
		// 1 because of "nick_msg" 1 white space
		msgLen := 5 + 3 + len(msg.Nick) + len(msg.Text) + 1

		// 1 because if msgLen < 40, the division gives 0 and one breaks another line for readability
		// nick > text
		// nick > text2
		msgLines := msgLen/40 + 1

		buf := minigo.GetMoveCursorAt(1, 24)
		for k := 0; k < msgLines; k += 1 {
			buf = append(buf, minigo.GetMoveCursorReturn(1)...)
		}
		buf = append(buf, minigo.GetMoveCursorAt(1, InputLine-msgLines-1)...)

		buf = append(buf, minigo.EncodeAttributes(minigo.InversionFond)...)
		buf = append(buf, minigo.EncodeSprintf("%s - %s", msg.Time.Format("15:04"), msg.Nick)...)
		buf = append(buf, minigo.EncodeAttributes(minigo.FondNormal)...)
		buf = append(buf, minigo.GetMoveCursorRight(1)...)

		if msg.Type == MessageTeletel {
			buf = append(buf, msg.Text...)
		} else {
			buf = append(buf, minigo.EncodeMessage(msg.Text)...)
		}
		m.Send(buf)
	}

	helpers(m)
}

func helpers(m *minigo.Minitel) {
	m.WriteHelperLeft(24, "MAJ ECRAN", "REPET.")
	m.WriteHelperRight(24, "MESSAGE +", "ENVOI")
}
