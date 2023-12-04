package minichat

import (
	"sync/atomic"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/prometheus/client_golang/prometheus"
)

func RunChatPage(m *minigo.Minitel, msgDB *databases.MessageDatabase, cntd *atomic.Int32, nick string, promMsgNb prometheus.Counter) (op int) {
	chatPage := minigo.NewPage("chat", m, nil)
	chatLayout := NewChatLayout(m, msgDB, cntd, nick)

	chatPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		m.RouleauOn()
		m.MinusculeOn()

		msgDB.Subscribe(nick)
		inputs.AppendInput("messages", minigo.NewInput(m, rowInput, 1, 40, 2, false))

		chatLayout.Init()
		inputs.InitAll()

		return minigo.NoOp
	})

	chatPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	chatPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		msgDB.Resign(nick)

		m.RouleauOff()
		m.MinusculeOff()
		return nil, minigo.SommaireOp
	})

	chatPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key int32) {
		inputs.AppendKeyActive(key)
	})

	_, op = chatPage.Run()
	return op
}
