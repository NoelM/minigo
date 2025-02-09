package superchat

import (
	"sync/atomic"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/NoelM/minigo/notel/logs"
	"github.com/prometheus/client_golang/prometheus"
)

func RunChatPage(m *minigo.Minitel, msgDB *databases.MessageDatabase, cntd *atomic.Int32, nick string, promMsgNb prometheus.Counter) (op int) {
	chatPage := minigo.NewPage("chat", m, nil)
	chatLayout := NewChatLayout(m, msgDB, cntd, nick)

	chatPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		m.Reset()

		m.RouleauOn()
		m.MinusculeOn()

		msgDB.Subscribe(nick)
		inputs.AppendInput("messages", minigo.NewInput(m, inputRow, 0, 39, 2, false))

		chatLayout.Init()
		inputs.InitAll()

		return minigo.NoOp
	})

	chatPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if len(inputs.ValueActive()) == 0 {
			return nil, minigo.NoOp
		}
		promMsgNb.Inc()

		msg := databases.Message{
			Nick: nick,
			Text: inputs.ValueActive(),
			Time: time.Now(),
		}
		msgDB.PushMessage(msg, false)

		logs.InfoLog("send new message to IRC from nick=%s len=%d\n", nick, len(msg.Text))

		inputs.HideAll()
		chatLayout.Update()

		inputs.ResetAll()
		return nil, minigo.NoOp
	})

	chatPage.SetRepetitionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.HideAll()
		chatLayout.Update()

		inputs.UnHideAll()
		return nil, minigo.NoOp
	})

	chatPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	chatPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		msgDB.Resign(nick)

		m.PrintStatus("")
		m.RouleauOff()
		m.MinusculeOff()
		return nil, minigo.SommaireOp
	})

	chatPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		chatLayout.PrintNextMessage()
		return nil, minigo.NoOp
	})

	chatPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		chatLayout.PrintPreviousMessage()
		return nil, minigo.NoOp
	})

	chatPage.SetGuideFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.GuideOp
	})

	chatPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key int32) {
		inputs.AppendKeyActive(key)
	})

	_, op = chatPage.Run()
	return op
}
