package superchat

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/NoelM/minigo/notel/metrics"
)

func ServiceSuperchat(minitel *minigo.Minitel, chatManager *databases.ChatManager, metrics *metrics.Metrics, nickname string) int {
	// First show channel selection page
CHANNEL:
	choice, op := ChannelPage(minitel, chatManager).Run()
	if op == minigo.SommaireOp {
		return op
	}

	selectedChannel := choice["channel"]
	if selectedChannel == "" {
		return minigo.SommaireOp
	}

	channel := chatManager.GetChannel(selectedChannel)
	if channel == nil {
		return minigo.SommaireOp
	}

HELP:
	_, op = HelpPage(minitel, channel).Run()
	if op != minigo.SommaireOp {
		return op
	}

	_, op = ChatPage(minitel, channel, metrics, nickname).Run()
	if op == minigo.GuideOp {
		goto HELP
	} else if op == minigo.SommaireOp {
		goto CHANNEL
	} else {
		return op
	}
}
