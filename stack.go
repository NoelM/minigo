package minigo

type Stack struct {
	maxSize   int
	curId     int
	container [][]byte
	empty     bool
}

func NewStack(maxSize int) *Stack {
	return &Stack{
		maxSize:   maxSize,
		container: make([][]byte, maxSize),
		curId:     0,
		empty:     true,
	}
}

func (s *Stack) InitPCE() *Stack {
	for id := range s.container {
		s.container[id] = make([]byte, 17)
	}

	return s
}

func (s *Stack) Reset() {
	s.curId = 0
	s.empty = true
}

func (s *Stack) Add(msg []byte) {
	s.empty = false

	s.container[s.curId] = make([]byte, len(msg))
	copy(s.container[s.curId], msg)

	s.curId += 1
	if s.curId == s.maxSize {
		s.curId = 0
	}
}

func (s *Stack) Last() []byte {
	return s.container[len(s.container)-1]
}

func (s *Stack) Get(id int) []byte {
	return s.container[id]
}

func (s *Stack) Empty() bool {
	return s.empty
}
