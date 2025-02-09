package superchat

import (
	"sync/atomic"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/prometheus/client_golang/prometheus"
)

func ServiceSuperchat(minitel *minigo.Minitel, msgDB *databases.MessageDatabase, cntd *atomic.Int32, nick string, promMsgNb prometheus.Counter) int {
HELP:
	_, op := HelpPage(minitel).Run()
	if op != minigo.SommaireOp {
		return op
	}

	op = RunChatPage(minitel, msgDB, cntd, nick, promMsgNb)
	if op == minigo.GuideOp {
		goto HELP
	} else {
		return op
	}
}
