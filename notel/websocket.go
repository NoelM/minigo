package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/confs"
	"github.com/NoelM/minigo/notel/logs"
	"nhooyr.io/websocket"
)

func serveWebSocket(wg *sync.WaitGroup, connConf confs.ConnectorConf, metrics *Metrics) {
	defer wg.Done()

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tagFull := fmt.Sprintf("%s:%s", connConf.Tag, r.RemoteAddr)

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
		if err != nil {
			logs.ErrorLog("[%s] serve-ws: unable to open websocket connection: %s\n", tagFull, err.Error())
			return
		}

		defer conn.Close(websocket.StatusInternalError, "websocket internal error, quitting")
		logs.InfoLog("[%s] serve-ws: new connection\n", tagFull)

		conn.SetReadLimit(1024)

		ctx, cancel := context.WithTimeout(r.Context(), time.Minute*10)
		defer cancel()

		ws, _ := minigo.NewWebsocket(conn, ctx)
		_ = ws.Init()

		var innerWg sync.WaitGroup
		innerWg.Add(2)

		network := minigo.NewNetwork(ws, false, &innerWg, "websocket")
		m := minigo.NewMinitel(network, false, &innerWg, connConf.Tag, metrics.ConnLostCount)
		m.NoCSI()
		go m.Serve()

		NotelApplication(m, &innerWg, &connConf, metrics)
		innerWg.Wait()

		logs.InfoLog("[%s] serve-ws: disconnect\n", tagFull)
		ws.Disconnect()

		logs.InfoLog("[%s] serve-ws: session closed\n", tagFull)
	})

	err := http.ListenAndServe(connConf.Path, fn)
	log.Fatal(err)
}
