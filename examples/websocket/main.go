package main

import (
	"log"
	"net/http"
	"os"

	"github.com/NoelM/minigo"
)

var infoLog = log.New(os.Stdout, "[MINICHAT] INFO:", log.Ldate|log.LUTC)
var warnLog = log.New(os.Stdout, "[MINICHAT] WARN:", log.Ldate|log.LUTC)
var errorLog = log.New(os.Stdout, "[MINICHAT] ERROR:", log.Ldate|log.LUTC)

func main() {

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ws, _ := minigo.NewWebsocket(w, r)
		err := ws.Init()
		if err != nil {
			errorLog.Printf("impossible to init websocket because: %s\n", err.Error())
		}

		m := minigo.NewMinitel(ws, false)
		go m.Listen()

		nick := logPage(m)

		ircDvr := NewIrcDriver(string(nick))
		go ircDvr.Loop()

		chatPage(m, ircDvr)
		ircDvr.Quit()

		infoLog.Printf("Minitel session closed for IP=%s nick=%s\n", r.RemoteAddr, nick)
	})

	err := http.ListenAndServe("192.168.1.34:3615", fn)
	log.Fatal(err)
}
