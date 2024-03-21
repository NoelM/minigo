package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/NoelM/minigo/notel/logs"
	"nhooyr.io/websocket"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var CommuneDb *databases.CommuneDatabase
var MessageDb *databases.MessageDatabase
var UsersDb *databases.UsersDatabase

var NbConnectedUsers atomic.Int32

var (
	promConnNb = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notel_connection_number",
		Help: "The total number connection to NOTEL",
	},
		[]string{"source"})

	promConnAttemptNb = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notel_connection_attempt_number",
		Help: "The total number of connection attempts to NOTEL",
	},
		[]string{"source"})

	promConnLostNb = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notel_connection_lost_number",
		Help: "The total number of lost connections on NOTEL",
	},
		[]string{"source"})

	promConnActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "notel_connection_active",
		Help: "The number of currently active connections to NOTEL",
	},
		[]string{"source"})

	promConnDur = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notel_connection_duration",
		Help: "The total connection duration to NOTEL",
	},
		[]string{"source"})

	promMsgNb = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notel_messages_number",
		Help: "The total number of postel messages to NOTEL",
	})
)

const (
	ServeWS              = true
	ServeUSR56FaxModem1  = true
	ServeUSR56KFaxModem2 = true
)

const (
	WSTag              = "ws"
	USR56KFaxModemTag1 = "usr-56k-faxmodem-1"
	USR56KFaxModemTag2 = "usr-56k-faxmodem-2"
)

func main() {
	var wg sync.WaitGroup

	CommuneDb = databases.NewCommuneDatabase()
	CommuneDb.LoadCommuneDatabase("/media/core/communes-departement-region.csv")

	MessageDb = databases.NewMessageDatabase()
	MessageDb.LoadMessages("/media/core/messages.db")

	UsersDb = databases.NewUsersDatabase()
	UsersDb.LoadDatabase("/media/core/users.db")

	if ServeWS {
		wg.Add(1)
		go serveWS(&wg, "192.168.1.34:3615")
	}

	if ServeUSR56FaxModem1 {
		wg.Add(1)
		go serveModem(&wg, ConfUSR56KFaxModem, "/dev/ttyUSB0", USR56KFaxModemTag1)
	}

	if ServeUSR56KFaxModem2 {
		wg.Add(1)
		go serveModem(&wg, ConfUSR56KFaxModem, "/dev/ttyUSB1", USR56KFaxModemTag2)
	}

	wg.Add(1)
	go serverMetrics(&wg)

	wg.Wait()

	MessageDb.Quit()
	UsersDb.Quit()
}

func serveWS(wg *sync.WaitGroup, url string) {
	defer wg.Done()

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tagFull := fmt.Sprintf("%s:%s", WSTag, r.RemoteAddr)

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

		m := minigo.NewMinitel(ws, false, WSTag, promConnLostNb, &innerWg)
		m.NoCSI()
		go m.Serve()

		NotelHandler(m, WSTag, &innerWg)
		innerWg.Wait()

		logs.InfoLog("[%s] serve-ws: disconnect\n", tagFull)
		ws.Disconnect()

		logs.InfoLog("[%s] serve-ws: session closed\n", tagFull)
	})

	err := http.ListenAndServe(url, fn)
	log.Fatal(err)
}

func serverMetrics(wg *sync.WaitGroup) {
	defer wg.Done()

	for _, cv := range []*prometheus.CounterVec{promConnNb, promConnLostNb, promConnDur, promConnAttemptNb} {
		cv.With(prometheus.Labels{"source": WSTag}).Inc()
		cv.With(prometheus.Labels{"source": USR56KFaxModemTag1}).Inc()
		cv.With(prometheus.Labels{"source": USR56KFaxModemTag2}).Inc()
	}
	promConnActive.With(prometheus.Labels{"source": WSTag}).Set(0)
	promConnActive.With(prometheus.Labels{"source": USR56KFaxModemTag1}).Set(0)
	promConnActive.With(prometheus.Labels{"source": USR56KFaxModemTag2}).Set(0)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func serveModem(wg *sync.WaitGroup, init []minigo.ATCommand, tty string, modemTag string) {
	defer wg.Done()

	modem, err := minigo.NewModem(tty, 115200, init, modemTag, promConnAttemptNb)
	if err != nil {
		log.Fatal(err)
	}

	err = modem.Init()
	if err != nil {
		log.Fatal(err)
	}

	modem.RingHandler(func(mdm *minigo.Modem) {
		var connectionWg sync.WaitGroup
		connectionWg.Add(2)

		m := minigo.NewMinitel(mdm, true, modemTag, promConnLostNb, &connectionWg)
		go m.Serve()

		NotelHandler(m, modemTag, &connectionWg)
		connectionWg.Wait()

		logs.InfoLog("[%s] ring-handler: disconnect\n", modemTag)
		mdm.Disconnect()

		logs.InfoLog("[%s] ring-handler: minitel session closed\n", modemTag)
	})

	modem.Serve(false)
}

func NotelHandler(mntl *minigo.Minitel, sourceTag string, wg *sync.WaitGroup) {

	promConnNb.With(prometheus.Labels{"source": sourceTag}).Inc()
	active := NbConnectedUsers.Add(1)
	promConnActive.With(prometheus.Labels{"source": sourceTag}).Inc()

	logs.InfoLog("[%s] notel-handler: start handler, connected=%d\n", sourceTag, active)
	startConn := time.Now()

	StartSplash(mntl)

SIGNIN:
	creds, op := NewPageSignIn(mntl).Run()

	if op == minigo.GuideOp {
		creds, op = NewSignUpPage(mntl).Run()

		if op == minigo.SommaireOp {
			goto SIGNIN
		}
	}

	if op == minigo.EnvoiOp {
		SommaireHandler(mntl, creds["login"])
	}

	promConnDur.With(prometheus.Labels{"source": sourceTag}).Add(time.Since(startConn).Seconds())

	active = NbConnectedUsers.Add(-1)
	promConnActive.With(prometheus.Labels{"source": sourceTag}).Dec()

	logs.InfoLog("[%s] notel-handler: quit handler, connected=%d\n", sourceTag, active)

	wg.Done()
}
