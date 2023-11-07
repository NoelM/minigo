package main

import (
	"context"
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
	promConnNb = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notel_connection_number",
		Help: "The total number connection to NOTEL",
	})

	promConnActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "notel_connection_active",
		Help: "The number of currently active connections to NOTEL",
	})

	promConnDur = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notel_connection_duration",
		Help: "The total connection duration to NOTEL",
	})

	promMsgNb = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notel_messages_number",
		Help: "The total number of postel messages to NOTEL",
	})
)

func main() {
	var wg sync.WaitGroup

	CommuneDb = NewCommuneDatabase()
	CommuneDb.LoadCommuneDatabase("/media/core/communes-departement-region.csv")

	MessageDb = NewMessageDatabase()
	MessageDb.LoadMessages("/media/core/messages.db")

	UsersDb = NewUsersDatabase()
	UsersDb.LoadDatabase("/media/core/users.db")

	wg.Add(4)

	go serveWS(&wg, "192.168.1.34:3615")

	USR56KPro := []minigo.ATCommand{
		{
			Command: "ATZ",
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
	go serveModem(&wg, USR56KPro, "/dev/ttyUSB0")

	USRSportster := []minigo.ATCommand{
		{
			Command: "ATZ",
			Reply:   "OK",
		},
		{
			Command: "AT&F1",
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
	go serveModem(&wg, USRSportster, "/dev/ttyUSB1")

	go serverMetrics(&wg)

	wg.Wait()

	MessageDb.Quit()
}

func serveWS(wg *sync.WaitGroup, url string) {
	defer wg.Done()

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
		if err != nil {
			errorLog.Printf("unable to open websocket connection: %s\n", err.Error())
			return
		}

		defer conn.Close(websocket.StatusInternalError, "websocket internal error, quitting")
		infoLog.Printf("new connection from IP=%s\n", r.RemoteAddr)

		conn.SetReadLimit(1024)

		ctx, cancel := context.WithTimeout(r.Context(), time.Minute*10)
		defer cancel()

		ws, _ := minigo.NewWebsocket(conn, ctx)
		_ = ws.Init()

		m := minigo.NewMinitel(ws, false)
		go m.Listen()

		NotelHandler(m)

		infoLog.Printf("Minitel session closed for IP=%s\n", r.RemoteAddr)
	})

	err := http.ListenAndServe(url, fn)
	log.Fatal(err)
}

func serverMetrics(wg *sync.WaitGroup) {
	defer wg.Done()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func serveModem(wg *sync.WaitGroup, init []minigo.ATCommand, tty string) {
	defer wg.Done()

	modem, err := minigo.NewModem(tty, 115200, init)
	if err != nil {
		log.Fatal(err)
	}

	err = modem.Init()
	if err != nil {
		log.Fatal(err)
	}

	modem.RingHandler(func(mdm *minigo.Modem) {
		m := minigo.NewMinitel(mdm, true)
		go m.Listen()

		NotelHandler(m)

		infoLog.Printf("Minitel session closed for Modem\n")
	})

	modem.Serve(false)
}

func NotelHandler(mntl *minigo.Minitel) {

	promConnNb.Inc()
	active := NbConnectedUsers.Add(1)
	promConnActive.Set(float64(active))

	infoLog.Printf("enters service handler, connected=%d\n", active)
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

	promConnDur.Add(time.Since(startConn).Seconds())

	active = NbConnectedUsers.Add(-1)
	promConnActive.Set(float64(active))

	infoLog.Printf("quits NOTEL service handler, connected=%d\n", active)
}
