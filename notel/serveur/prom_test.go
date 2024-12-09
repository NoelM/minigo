package serveur

import (
	"fmt"
	"testing"
)

func TestRequest(t *testing.T) {
	r, _ := Request(ConnectWeekly)
	fmt.Println(r.Data)

	r, _ = Request(DurationWeekly)
	fmt.Println(r.Data)

	r, _ = Request(MessagesWeekly)
	fmt.Println(r.Data)

	r, _ = Request(CPULoad)
	fmt.Println(r.Data)

	r, _ = Request(CPUTemp)
	fmt.Println(r.Data)

}
