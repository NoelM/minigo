package serveur

import "testing"

func TestRequest(t *testing.T) {
	Request(ConnectWeekly)
	Request(DurationWeekly)
	Request(MessagesWeekly)
	Request(CPULoad)
	Request(CPUTemp)
}
