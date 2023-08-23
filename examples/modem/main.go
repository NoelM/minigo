package main

import (
	"log"
	"os"

	"github.com/NoelM/minigo"
)

var infoLog = log.New(os.Stdout, "[minichat] info:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var warnLog = log.New(os.Stdout, "[minichat] warn:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var errorLog = log.New(os.Stdout, "[minichat] error:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

func main() {
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

	modem := minigo.NewModem("/dev/ttyUSB0", 115200, init)

	err := modem.Init()
	if err != nil {
		log.Fatal(err)
	}

	modem.RingHandler(func(mdm *minigo.Modem) {
		m := minigo.NewMinitel(mdm, true)
		go m.Listen()

		nick := logPage(m)
		if !mdm.Connected() || len(nick) == 0 {
			return
		}

		ircDvr := NewIrcDriver(string(nick))
		go ircDvr.Loop()

		chatPage(m, ircDvr)
		ircDvr.Quit()

		infoLog.Printf("Minitel session closed from Modem nick=%s\n", nick)
	})

	modem.Serve(false)
}
