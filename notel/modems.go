package main

import "github.com/NoelM/minigo"

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
<<<<<<< Updated upstream
	{
		Command: "AT&N2",
		Reply:   "OK",
	},
	// S27=16: V23 mode enabled
	{
		Command: "ATS27=16",
=======
	// S27=16: fallback on V.23
	{
		Command: "ATE0L0M0X4&A0&N2S0=1S27=16",
>>>>>>> Stashed changes
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
