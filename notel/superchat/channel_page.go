package superchat

import (
	"strconv"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/NoelM/minigo/notel/logs"
)

func ChannelPage(m *minigo.Minitel, chatManager *databases.ChatManager) *minigo.Page {
	channelPage := minigo.NewPage("channel", m, nil)

	channelPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		m.Reset()
		m.CursorOff()
		m.SendVDT("static/superchat.vdt")

		// Display header
		m.ModeG0()
		m.Attributes(minigo.FondNoir, minigo.CaractereBlanc, minigo.GrandeurNormale)

		mntl.MoveAt(7, 1)
		mntl.Attributes(minigo.DoubleHauteur)
		mntl.Print("Salons disponibles")
		mntl.Attributes(minigo.GrandeurNormale)

		// Display available channels
		channels := chatManager.ListChannels()
		names := make([]string, 0, len(channels))
		for _, ch := range channels {
			names = append(names, ch.Name)
		}
		list := minigo.NewListEnum(mntl, names, 9, 1, 22, 2)
		list.Display()

		mntl.MoveAt(24, 0)
		mntl.HelperRight("SALON: .. +", "ENVOI", minigo.FondVert, minigo.CaractereNoir)
		inputs.AppendInput("channel", minigo.NewInput(mntl, 24, 27, 2, 1, true))

		inputs.InitAll()
		return minigo.NoOp
	})

	channelPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		channelId, err := strconv.Atoi(inputs.ValueActive())
		if err != nil {
			return nil, minigo.NoOp
		}

		channelSlug := chatManager.ListChannels()[channelId-1].Slug
		logs.InfoLog("selected channel: %s\n", channelSlug)

		return map[string]string{"channel": channelSlug}, minigo.EnvoiOp
	})

	channelPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	channelPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	channelPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key int32) {
		inputs.AppendKeyActive(key)
	})

	return channelPage
}
