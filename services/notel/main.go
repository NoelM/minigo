package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NoelM/minigo"
	"nhooyr.io/websocket"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var infoLog = log.New(os.Stdout, "[notel] info:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var warnLog = log.New(os.Stdout, "[notel] warn:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var errorLog = log.New(os.Stdout, "[notel] error:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

var CommuneDb *CommuneDatabase
var MessageDb *MessageDatabase
var UsersDb *UsersDatabase

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
	ServeWS             = true
	ServeUSR56KPro      = false
	ServeUSR56KFaxModem = true
)

const (
	WSTag             = "ws"
	USR56KProTag      = "usr-56k-pro"
	USR56KFaxModemTag = "usr-56k-faxmodem"
)

func main() {
	var wg sync.WaitGroup

	CommuneDb = NewCommuneDatabase()
	CommuneDb.LoadCommuneDatabase("/media/core/communes-departement-region.csv")

	MessageDb = NewMessageDatabase()
	MessageDb.LoadMessages("/media/core/messages.db")

	UsersDb = NewUsersDatabase()
	UsersDb.LoadDatabase("/media/core/users.db")

	if ServeWS {
		wg.Add(1)
		go serveWS(&wg, "192.168.1.34:3615")
	}

	if ServeUSR56KPro {
		USR56KPro := []minigo.ATCommand{
			{
				Command: "ATZ0",
				Reply:   "OK",
			},
			{
				Command: "AT&F1+MCA=0",
				Reply:   "OK",
			},
			{
				Command: "ATL0M0",
				Reply:   "OK",
			},
			{
				Command: "AT&N2",
				Reply:   "OK",
			},
			{
				Command: "ATS27=16",
				Reply:   "OK",
			},
		}
		wg.Add(1)
		go serveModem(&wg, USR56KPro, "/dev/ttyUSB0", USR56KProTag)
	}

	if ServeUSR56KFaxModem {
		USR56KFaxModem := []minigo.ATCommand{
			{
				Command: "ATM0L0E0&H1&S1&R2",
				Reply:   "OK",
			},
			{
				Command: "ATS27=16S34=8S9=100&B1",
				Reply:   "OK",
			},
		}
		wg.Add(1)
		go serveModem(&wg, USR56KFaxModem, "/dev/ttyUSB0", USR56KFaxModemTag)
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
			errorLog.Printf("[%s] serve-ws: unable to open websocket connection: %s\n", tagFull, err.Error())
			return
		}

		defer conn.Close(websocket.StatusInternalError, "websocket internal error, quitting")
		infoLog.Printf("[%s] serve-ws: new connection\n", tagFull)

		conn.SetReadLimit(1024)

		ctx, cancel := context.WithTimeout(r.Context(), time.Minute*10)
		defer cancel()

		ws, _ := minigo.NewWebsocket(conn, ctx)
		_ = ws.Init()

		var innerWg sync.WaitGroup
		innerWg.Add(2)

		m := minigo.NewMinitel(ws, false, WSTag, promConnLostNb, &innerWg)
		go m.Listen()

		NotelHandler(m, WSTag, &innerWg)
		innerWg.Wait()

		infoLog.Printf("[%s] serve-ws: disconnect\n", tagFull)
		ws.Disconnect()

		infoLog.Printf("[%s] serve-ws: session closed\n", tagFull)
	})

	err := http.ListenAndServe(url, fn)
	log.Fatal(err)
}

func serverMetrics(wg *sync.WaitGroup) {
	defer wg.Done()

	for _, cv := range []*prometheus.CounterVec{promConnNb, promConnLostNb, promConnDur, promConnAttemptNb} {
		cv.With(prometheus.Labels{"source": WSTag}).Inc()
		cv.With(prometheus.Labels{"source": USR56KProTag}).Inc()
		cv.With(prometheus.Labels{"source": USR56KFaxModemTag}).Inc()
	}
	promConnActive.With(prometheus.Labels{"source": WSTag}).Set(0)
	promConnActive.With(prometheus.Labels{"source": USR56KProTag}).Set(0)
	promConnActive.With(prometheus.Labels{"source": USR56KFaxModemTag}).Set(0)

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
		go m.Listen()

		NotelHandler(m, modemTag, &connectionWg)
		connectionWg.Wait()

		infoLog.Printf("[%s] ring-handler: disconnect\n", modemTag)
		mdm.Disconnect()

		infoLog.Printf("[%s] ring-handler: minitel session closed\n", modemTag)
	})

	modem.Serve(false)
}

func NotelHandler(mntl *minigo.Minitel, sourceTag string, wg *sync.WaitGroup) {

	promConnNb.With(prometheus.Labels{"source": sourceTag}).Inc()
	active := NbConnectedUsers.Add(1)
	promConnActive.With(prometheus.Labels{"source": sourceTag}).Set(float64(active))

	infoLog.Printf("[%s] notel-handler: start handler, connected=%d\n", sourceTag, active)
	startConn := time.Now()

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
	promConnActive.With(prometheus.Labels{"source": sourceTag}).Set(float64(active))

	infoLog.Printf("[%s] notel-handler: quit handler, connected=%d\n", sourceTag, active)

	wg.Done()
}
