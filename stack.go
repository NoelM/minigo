package minigo

type Stack struct {
	maxSize   int
	lastId    int
	container [][]byte
}

func NewStack(maxSize int) *Stack {
	return &Stack{
		maxSize:   maxSize,
		container: make([][]byte, 0),
	}
}

func (s *Stack) shift() {
	s.container = s.container[1:]
}

func (s *Stack) Add(msg []byte) {
	if len(s.container) == s.maxSize {
		s.shift()
	}

	buf := make([]byte, len(msg))
	copy(buf, msg)

	s.container = append(s.container, buf)
	s.lastId += 1
}

func (s *Stack) Last() []byte {
	if len(s.container) == 0 {
		return nil
	}

	return s.container[len(s.container)-1]
}

func (s *Stack) Get(id int) []byte {
	if len(s.container) == 0 {
		return nil
	}

	if id < s.lastId-s.maxSize {
		return nil
	}

	relPos := (len(s.container) - 1) - (s.lastId - id)
	return s.container[relPos]
}
