package main

import (
	"fmt"
	"time"

	"github.com/NoelM/minigo"
)

func ServiceMiniChat(m *minigo.Minitel, login string) int {
	if _, serviceId := NewChatPage(m, login).Run(); serviceId != minigo.NoOp {
		return serviceId
	}

	infoLog.Printf("minichat session closed for nick=%s\n", login)

	return sommaireId
}

func NewChatPage(m *minigo.Minitel, nickname string) *minigo.Page {
	chatPage := minigo.NewPage("chat", m, nil)
	lastMessageDate := time.Time{}

	chatPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		infoLog.Printf("opening chat page for nick=%s\n", nickname)

		MessageDb.Subscribe(nickname)
		inputs.AppendInput("messages", minigo.NewInput(m, InputLine, 1, 40, 2, false))

		m.RouleauOn()
		m.MinusculeOn()
		updateScreen(m, nickname, &lastMessageDate)

		inputs.InitAll()
		return minigo.NoOp
	})

	chatPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if len(inputs.ValueActive()) == 0 {
			return nil, minigo.NoOp
		}
		promMsgNb.Inc()

		msg := Message{
			Nick: nickname,
			Text: inputs.ValueActive(),
			Time: time.Now(),
		}
		MessageDb.PushMessage(msg, false)
		//ircDrv.SendMessage(msg)

		infoLog.Printf("send new message to IRC from nick=%s len=%d\n", nickname, len(msg.Text))

		inputs.HideAll()
		updateScreen(m, nickname, &lastMessageDate)

		inputs.ResetAll()

		return nil, minigo.NoOp
	})

	chatPage.SetRepetitionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		infoLog.Printf("user nick=%s asked for a refresh\n", nickname)
		if !MessageDb.HasNewMessage(nickname) {
			return nil, minigo.NoOp
		}

		inputs.HideAll()
		updateScreen(m, nickname, &lastMessageDate)
		inputs.UnHideAll()

		return nil, minigo.NoOp
	})

	chatPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	chatPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		MessageDb.Resign(nickname)
		m.RouleauOff()
		m.MinusculeOff()
		return nil, sommaireId
	})

	chatPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key int32) {
		inputs.AppendKeyActive(key)
	})

	return chatPage
}

const InputLine = 22

func updateScreen(m *minigo.Minitel, nick string, lastMessageDate *time.Time) {
	m.CursorOff()

	// Clean helpers
	m.MoveCursorAt(24, 1)
	m.CleanLine()

	// Get all the messages from the DB
	lastMessages := MessageDb.GetMessages(nick)

	// Display only the messages that fit in screen
	// line=1           -- Message Zone
	//  ...             --  ...
	// line=InputLine-1 -- End of Message Zone
	nbLines := 0
	firstMsgId := 0
	localLastDate := *lastMessageDate
	for id := len(lastMessages) - 1; id >= 0; id -= 1 {
		nbLines += len(lastMessages[id].Text)/minigo.ColonnesSimple + 1

		if getDateString(localLastDate, lastMessages[id].Time) != "" {
			nbLines += 3
		}
		localLastDate = lastMessages[id].Time

		if nbLines >= InputLine-1 {
			firstMsgId = id
			break
		}
	}

	// Get only the displayable messages
	lastMessages = lastMessages[firstMsgId:]

	// Print those messages with date separators if needed
	for _, msg := range lastMessages {
		printDate(m, *lastMessageDate, msg.Time)
		*lastMessageDate = msg.Time

		printOneMsg(m, msg)
	}

	printChatHelpers(m)
}

func printDate(m *minigo.Minitel, lastDate time.Time, date time.Time) {
	dateString := getDateString(lastDate, date)
	if dateString == "" {
		return
	}

	buf := minigo.GetMoveCursorAt(24, 1)
	// this is not a repetition
	// needed in rouleau mode
	buf = append(buf, minigo.GetMoveCursorReturn(1)...)
	buf = append(buf, minigo.GetMoveCursorReturn(1)...)
	m.Send(buf)

	m.WriteStringCenter(InputLine-2, dateString)
}

func getDateString(lastDate time.Time, date time.Time) (dateString string) {
	durationSinceLastMsg := date.Sub(lastDate)

	if durationSinceLastMsg >= 365*24*time.Hour {
		dateString = fmt.Sprintf("%s %d %s %d, %s",
			weekdayIdToString(date.Weekday()),
			date.Day(),
			monthIdToString(date.Month()),
			date.Year(),
			date.Format("15:04"))
	} else if durationSinceLastMsg >= 24*time.Hour || date.Day() != lastDate.Day() {
		dateString = fmt.Sprintf("%s %d %s, %s",
			weekdayIdToString(date.Weekday()),
			date.Day(),
			monthIdToString(date.Month()),
			date.Format("15:04"))
	} else if durationSinceLastMsg > 10*time.Minute {
		dateString = date.Format("15:04")
	}

	return
}

func printOneMsg(m *minigo.Minitel, msg Message) {
	// Message Format
	// [nick]>_[msg]
	// 2 because of ">_"
	msgLen := len(msg.Nick) + 2 + len(msg.Text)

	// 1 because if msgLen < 40, the division gives 0 and one breaks another line for readability
	// nick > text
	// nick > text2
	msgLines := msgLen/40 + 1

	// Rouleau mode, push to the top the messages
	buf := minigo.GetMoveCursorAt(24, 1)
	for k := 0; k < msgLines; k += 1 {
		buf = append(buf, minigo.GetMoveCursorReturn(1)...)
	}
	buf = append(buf, minigo.GetMoveCursorAt(InputLine-msgLines-1, 1)...)

	// Print nickname
	buf = append(buf, minigo.EncodeSprintf("%s> ", msg.Nick)...)

	// Print Message
	buf = append(buf, minigo.EncodeMessage(msg.Text)...)

	m.Send(buf)
}

func printChatHelpers(m *minigo.Minitel) {
	m.WriteHelperLeft(24, "MAJ ECRAN", "REPET.")
	m.WriteHelperRight(24, "MESSAGE +", "ENVOI")
}
