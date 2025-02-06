package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/confs"
	"github.com/NoelM/minigo/notel/logs"
	"github.com/reiver/go-telnet"
)

type notelHandler struct {
	connConf confs.ConnectorConf
	metrics  *Metrics
}

func (h notelHandler) ServeTELNET(ctx telnet.Context, w telnet.Writer, r telnet.Reader) {

	tagFull := fmt.Sprintf("%s", h.connConf.Tag)

	tel, _ := minigo.NewTelnet(ctx, w, r)
	_ = tel.Init()

	var innerWg sync.WaitGroup
	innerWg.Add(2)

	network := minigo.NewNetwork(tel, false, &innerWg, "telnet")
	m := minigo.NewMinitel(network, false, &innerWg, h.connConf.Tag, h.metrics.ConnLostCount)
	m.NoCSI()
	go m.Serve()

	NotelApplication(m, &innerWg, &h.connConf, h.metrics)
	innerWg.Wait()

	logs.InfoLog("[%s] serve-telnet: disconnect\n", tagFull)
	tel.Disconnect()

	logs.InfoLog("[%s] serve-telnet: session closed\n", tagFull)
}

func serveTelnet(wg *sync.WaitGroup, connConf confs.ConnectorConf, metrics *Metrics) {
	defer wg.Done()

	var fn telnet.Handler = notelHandler{
		connConf: connConf,
		metrics:  metrics,
	}
	err := telnet.ListenAndServe(connConf.Path, fn)
	log.Fatal(err)
}
