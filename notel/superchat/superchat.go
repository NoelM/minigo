package superchat

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/NoelM/minigo/notel/metrics"
)

func ServiceSuperchat(minitel *minigo.Minitel, messagesDb *databases.MessageDatabase, metrics *metrics.Metrics, nickname string) int {
HELP:
	_, op := HelpPage(minitel, metrics).Run()
	if op != minigo.SommaireOp {
		return op
	}

	op = RunChatPage(minitel, messagesDb, metrics, nickname)
	if op == minigo.GuideOp {
		goto HELP
	} else {
		return op
	}
}
