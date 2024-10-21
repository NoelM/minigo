package main

import (
	"log"
	"sync"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/confs"
	"github.com/NoelM/minigo/notel/logs"
)

var ConfUSR56KFaxModem = []minigo.ATCommand{
	// Z0: Reset configuration
	{
		Command: "ATZ0",
		Reply:   "OK",
	},
	// X4:  Full length modem reply
	// M0:  Speaker OFF
	// L0:  Speaker volume LOW
	// E0:  No command echo
	// &N2:    1200 bps connection default
	// S27=16: fallback on V.23
	{
		Command: "ATE0L0M0X4&A0&N2S0=1S27=16",
		Reply:   "OK",
	},
}

/*
var ConfUSR56KFaxModem = []minigo.ATCommand{
	// Z0: Reset configuration
	{
		Command: "ATZ0",
		Reply:   "OK",
	},
	// X4:  Full length modem reply
	// M0:  Speaker OFF
	// L0:  Speaker volume LOW
	// E0:  No command echo
	// &H1: Hardware control flow, Clear to Send (CTS)
	// &S1: Data Send Ready always ON
	// &R2: Recieved Data to computer only on RTS
	{
		Command: "ATX4M0L0E0&H1&S1&R2",
		Reply:   "OK",
	},
	// &N2:    1200 bps connection default
	// S27=16: V23 mode enabled
	// S9=6:   Duration of remote modem duration carrier recognition
	//         in tenth of seconds (here 60s)
	// &B1:    Fixed serial port rate
	{
		Command: "ATS27=16S9=6&B1",
		Reply:   "OK",
	},
}
*/

func serveModem(wg *sync.WaitGroup, connConf confs.ConnectorConf) {
	defer wg.Done()

	modem, err := minigo.NewModem(connConf.Path, 115200, connConf.Config, connConf.Tag, promConnAttemptNb)
	if err != nil {
		log.Fatal(err)
	}

	err = modem.Init()
	if err != nil {
		log.Fatal(err)
	}

	modem.RingHandler(func(modem *minigo.Modem) {
		var group sync.WaitGroup

		network := minigo.NewNetwork(modem, true, &group, connConf.Tag)
		minitel := minigo.NewMinitel(network, true, connConf.Tag, promConnLostNb, &group)
		go minitel.Serve()

		NotelApplication(minitel, connConf.Tag, &group)
		group.Wait()

		logs.InfoLog("[%s] ring-handler: disconnect\n", connConf.Tag)
		modem.Disconnect()

		logs.InfoLog("[%s] ring-handler: minitel session closed\n", connConf.Tag)
	})

	modem.Serve(false)
}
