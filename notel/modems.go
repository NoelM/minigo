package main

import "github.com/NoelM/minigo"

var ConfUSR56KPro = []minigo.ATCommand{
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

var ConfUSR56KFaxModem = []minigo.ATCommand{
	{
		Command: "ATM0L0E0&H1&S1&R2",
		Reply:   "OK",
	},
	{
		Command: "ATS27=16S34=8S9=100&B1",
		Reply:   "OK",
	},
}
