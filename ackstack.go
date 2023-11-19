package minigo

type AckType uint

const (
	NoAck = iota
	AckRouleau
	AckPage
	AckMinuscule
	AckMajuscule
	AckPCEStart
	AckPCEStop
)

type AckStack struct {
	container []AckType
}

func NewAckStack() *AckStack {
	return &AckStack{container: make([]AckType, 0)}
}

func (a *AckStack) Add(ack AckType) {
	a.container = append(a.container, ack)
}

func (a *AckStack) Pop() (ack AckType, ok bool) {
	if len(a.container) == 0 {
		return 0, false
	}

	ack = a.container[0]
	a.container = a.container[1:]

	return ack, true
}

func (a *AckStack) Len() int {
	return len(a.container)
}

func (a *AckStack) Empty() bool {
	return len(a.container) == 0
}
