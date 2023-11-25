package minigo

type Stack struct {
	container [][]byte
}

func NewStack() *Stack {
	return &Stack{container: make([][]byte, 0)}
}

func (s *Stack) Add(data ...[]byte) {
	for _, d := range data {
		item := make([]byte, len(data))
		copy(item, d)

		s.container = append(s.container, item)
	}
}

func (s *Stack) Pop() []byte {
	if len(s.container) == 0 {
		return nil
	}

	data := make([]byte, len(s.container[0]))
	copy(data, s.container[0])

	s.container = s.container[1:]

	return data
}

func (s *Stack) Empty() bool {
	return len(s.container) == 0
}
