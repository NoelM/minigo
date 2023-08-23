package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/NoelM/minigo"
	"nhooyr.io/websocket"
)

var infoLog = log.New(os.Stdout, "[notel] info:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var warnLog = log.New(os.Stdout, "[notel] warn:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var errorLog = log.New(os.Stdout, "[notel] error:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

func main() {
	var wg sync.WaitGroup

	go serveWS(&wg)
	go serveModem(&wg)

	wg.Wait()
}

func serveWS(wg *sync.WaitGroup) {
	wg.Add(1)

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

		ServiceHandler(m)

		infoLog.Printf("Minitel session closed for IP=%s\n", r.RemoteAddr)
	})

	err := http.ListenAndServe("192.168.1.34:3615", fn)
	log.Fatal(err)

	wg.Done()
}

func serveModem(wg *sync.WaitGroup) {
	wg.Add(1)

	init := []minigo.ATCommand{
		{
			Command: "AT&F1+MCA=0",
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

	modem, err := minigo.NewModem("/dev/ttyUSB0", 115200, init)
	if err != nil {
		log.Fatal(err)
	}

	err = modem.Init()
	if err != nil {
		log.Fatal(err)
	}

	modem.RingHandler(func(mdm *minigo.Modem) {
		m := minigo.NewMinitel(mdm, false)
		go m.Listen()

		ServiceHandler(m)

		infoLog.Printf("Minitel session closed for Modem\n")
	})

	modem.Serve(false)

	wg.Done()
}

func ServiceHandler(m *minigo.Minitel) {
	var id int
	for id >= sommaireId {
		switch id {
		case sommaireId:
			id = PageSommaire(m)
		case ircId:
			id = ServiceMiniChat(m)
		default:
			continue
		}
	}
}
