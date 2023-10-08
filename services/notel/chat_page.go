package main

import (
	"fmt"
	"time"

	"github.com/NoelM/minigo"
)

func ServiceMiniChat(m *minigo.Minitel) int {
	out, serviceId := NewLogPage(m).Run()
	nick, ok := out["nick"]

	if len(nick) == 0 || !ok {
		return sommaireId
	} else if serviceId != minigo.NoOp && serviceId != minigo.QuitPageOp {
		return serviceId
	}

	ircDvr := NewIrcDriver(string(nick))
	go ircDvr.Loop()

	_, serviceId = NewChatPage(m, ircDvr).Run()
	ircDvr.Quit()

	if serviceId != minigo.NoOp {
		return serviceId
	}

	infoLog.Printf("minichat session closed for nick=%s\n", out)

	return sommaireId
}

func NewChatPage(m *minigo.Minitel, ircDrv *IrcDriver) *minigo.Page {
	chatPage := minigo.NewPage("chat", m, nil)
	lastMessageDate := time.Time{}
	nickname := ""

	chatPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		nickname = ircDrv.Nick
		infoLog.Printf("opening chat page for nick=%s\n", nickname)

		MessageDb.Subscribe(nickname)
		inputs.AppendInput("messages", minigo.NewInput(m, 1, InputLine, 40, 5, ">", true))

		m.RouleauOn()
		updateScreen(m, nickname, &lastMessageDate)

		inputs.RepetitionActive()

		return minigo.NoOp
	})

	chatPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		msg := Message{
			Nick: ircDrv.Nick,
			Text: string(inputs.ValueActive()),
			Time: time.Now(),
		}
		MessageDb.PushMessage(msg, false)
		ircDrv.SendMessage(msg)

		infoLog.Printf("send new message to IRC from nick=%s len=%d\n", ircDrv.Nick, len(msg.Text))

		inputs.ClearActive()
		updateScreen(m, nickname, &lastMessageDate)

		inputs.RepetitionActive()

		return nil, minigo.NoOp
	})

	chatPage.SetRepetitionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		infoLog.Printf("user nick=%s asked for a refresh\n", ircDrv.Nick)

		inputs.ClearScreenAll()
		updateScreen(m, nickname, &lastMessageDate)
		inputs.RepetitionAll()

		return nil, minigo.NoOp
	})

	chatPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	chatPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		MessageDb.Resign(nickname)
		m.RouleauOff()
		return nil, sommaireId
	})

	chatPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key uint) {
		inputs.AppendKeyActive(byte(key))
	})

	return chatPage
}

const InputLine = 22

func updateScreen(m *minigo.Minitel, nick string, lastMessageDate *time.Time) {
	m.CursorOff()

	// Get all the messages from the DB
	lastMessages := MessageDb.GetMessages(nick)

	// Display only the messages that fit in screen
	// line=1           -- Message Zone
	//  ...             --  ...
	// line=InputLine-1 -- End of Message Zone
	nbLines := 0
	firstMsgId := 0
	for id := len(lastMessages) - 1; id >= 0; id -= 1 {
		nbLines += len(lastMessages[id].Text)/minigo.ColonnesSimple + 1
		if nbLines >= InputLine-1 {
			firstMsgId = id
			break
		}
	}

	// Get only the displayable messages
	lastMessages = lastMessages[firstMsgId:]

	// Print those messages with date separators if needed
	for _, msg := range lastMessages {
		printDate(m, lastMessageDate, msg.Time)
		*lastMessageDate = msg.Time

		printOneMsg(m, msg)
	}

	printChatHelpers(m)
}

func printDate(m *minigo.Minitel, lastDate *time.Time, date time.Time) {
	durationSinceLastMsg := date.Sub(*lastDate)

	dateString := ""
	if durationSinceLastMsg >= 365*24*time.Hour {
		dateString = fmt.Sprintf("%s %d %s %d, %s",
			weekdayIdToString(date.Weekday()),
			date.Day(),
			monthIdToString(date.Month()),
			date.Year(),
			date.Format("15:04"))
	} else if durationSinceLastMsg >= 24*time.Hour {
		dateString = fmt.Sprintf("%s %d %s, %s",
			weekdayIdToString(date.Weekday()),
			date.Day(),
			monthIdToString(date.Month()),
			date.Format("15:04"))
	} else if durationSinceLastMsg > 10*time.Minute {
		dateString = date.Format("15:04")
	} else {
		return
	}

	buf := minigo.GetMoveCursorAt(1, 24)
	// this is not a repetition
	// needed in rouleau mode
	buf = append(buf, minigo.GetMoveCursorReturn(1)...)
	buf = append(buf, minigo.GetMoveCursorReturn(1)...)
	m.Send(buf)

	m.WriteStringCenter(InputLine-2, dateString)
}

func printOneMsg(m *minigo.Minitel, msg Message) {
	// Message Format
	// [nick]> [msg]
	// 2 because of "> "
	msgLen := len(msg.Nick) + 2 + len(msg.Text)

	// 1 because if msgLen < 40, the division gives 0 and one breaks another line for readability
	// nick > text
	// nick > text2
	msgLines := msgLen/40 + 1

	buf := minigo.GetMoveCursorAt(1, 24)
	for k := 0; k < msgLines; k += 1 {
		buf = append(buf, minigo.GetMoveCursorReturn(1)...)
	}
	buf = append(buf, minigo.GetMoveCursorAt(1, InputLine-msgLines-1)...)
	buf = append(buf, minigo.EncodeSprintf("%s> ", msg.Nick)...)
	buf = append(buf, minigo.EncodeMessage(msg.Text)...)

	m.Send(buf)
}

func printChatHelpers(m *minigo.Minitel) {
	m.WriteHelperLeft(24, "MAJ ECRAN", "REPET.")
	m.WriteHelperRight(24, "MESSAGE +", "ENVOI")
}
