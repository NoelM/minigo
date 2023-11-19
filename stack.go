package minigo

type Stack struct {
	maxSize   int
	curId     int
	container [][]byte
}

func NewStack(maxSize int) *Stack {
	return &Stack{
		maxSize:   maxSize,
		container: make([][]byte, maxSize),
		curId:     0,
	}
}

func (s *Stack) InitPCE() *Stack {
	for id := range s.container {
		s.container[id] = make([]byte, 17)
	}

	return s
}

func (s *Stack) Add(msg []byte) {
	buf := make([]byte, len(msg))
	copy(buf, msg)

	s.container[s.curId] = buf

	s.curId += 1
	if s.curId%s.maxSize == 0 {
		s.curId = 0
	}
}

func (s *Stack) Last() []byte {
	return s.container[len(s.container)-1]
}

func (s *Stack) Get(id int) []byte {
	return s.container[id]
}
