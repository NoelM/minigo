package minigo

import "fmt"

type Matrix struct {
	inputs []*Input
	rows   int
	cols   int
	active int
}

func NewMatrix(rows, cols int) *Matrix {
	return &Matrix{
		inputs: make([]*Input, rows*cols),
		rows:   rows,
		cols:   cols,
	}
}

func (m *Matrix) SetInput(row, col int, input *Input) {
	if row > m.rows-1 || col > m.cols-1 {
		return
	}
	m.inputs[row*m.cols+col] = input
}

func (m *Matrix) ToMap() map[string]string {
	out := make(map[string]string)
	for id := range m.inputs {
		if in := m.inputs[id]; in != nil {
			out[fmt.Sprintf("%d", id)] = string(in.Value)
		}
	}

	return out
}

func (m *Matrix) ToArray() []string {
	values := make([]string, len(m.inputs))
	for id, in := range m.inputs {
		if in == nil {
			continue
		}
		values[id] = string(in.Value)
	}
	return values
}

func (m *Matrix) ValueActive() string {
	if len(m.inputs) == 0 {
		return ""
	}

	if in := m.inputs[m.active]; in != nil {
		return string(in.Value)
	} else {
		return ""
	}
}

func (m *Matrix) InitAll() {
	for id := range m.inputs {
		if in := m.inputs[id]; in != nil {
			in.Init()
		}
	}

	m.ActivateFirst()
}

func (m *Matrix) AppendKeyActive(key rune) {
	if len(m.inputs) == 0 {
		return
	}
	m.inputs[m.active].AppendKey(key)
}

func (m *Matrix) CorrectionActive() {
	if len(m.inputs) == 0 {
		return
	}
	m.inputs[m.active].Correction()
}

func (m *Matrix) UnHideActive() {
	if len(m.inputs) == 0 {
		return
	}
	m.inputs[m.active].UnHide()
}

func (m *Matrix) UnHideAll() {
	for _, in := range m.inputs {
		if in != nil {
			in.UnHide()
		}
	}
	m.activateInput()
}

func (m *Matrix) HideActive() {
	if len(m.inputs) == 0 {
		return
	}
	m.inputs[m.active].Hide()
}

func (m *Matrix) HideAll() {
	for _, in := range m.inputs {
		if in != nil {
			in.Hide()
		}
	}
}

func (m *Matrix) ResetActive() {
	if len(m.inputs) == 0 {
		return
	}
	m.inputs[m.active].Reset()
}

func (m *Matrix) ResetAll() {
	for _, in := range m.inputs {
		if in != nil {
			in.Reset()
		}
	}
	m.ActivateFirst()
}

func (m *Matrix) activateInput() {
	if len(m.inputs) == 0 {
		return
	}
	m.inputs[m.active].Activate()
}

func (m *Matrix) ActivateFirst() {
	for id, in := range m.inputs {
		if in != nil {
			m.active = id
			in.Activate()
			break
		}
	}
}

func (m *Matrix) ActivateLeft() {
	fmt.Println("activate left")
	nextActive := m.active

	for i := m.active; i >= 0; i -= 1 {
		fmt.Println("test", i)
		if m.inputs[i] != nil {
			nextActive = i
			break
		}
	}

	fmt.Println("new active", nextActive)
	m.active = nextActive
	m.activateInput()
}

func (m *Matrix) ActivatePrev() {
	m.ActivateLeft()
}

func (m *Matrix) ActivateRight() {
	fmt.Println("activate right")
	nextActive := m.active

	for i := m.active; i < len(m.inputs); i += 1 {
		fmt.Println("test", i)
		if m.inputs[i] != nil {
			nextActive = i
			break
		}
	}

	fmt.Println("new active", nextActive)
	m.active = nextActive
	m.activateInput()
}

func (m *Matrix) ActivateNext() {
	m.ActivateRight()
}

func (m *Matrix) ActivateUp() {
	nextActive := m.active

	upPos := m.active - m.cols
	for i := 0; i < len(m.inputs); i += 1 {
		if upPos-i >= 0 && upPos-i < len(m.inputs) && m.inputs[upPos-i] != nil {
			nextActive = upPos - i
			break
		}
		if upPos+i >= 0 && upPos+i < len(m.inputs) && m.inputs[upPos+i] != nil {
			nextActive = upPos + i
			break
		}
	}

	m.active = nextActive
	m.activateInput()
}

func (m *Matrix) ActivateDown() {
	nextActive := m.active

	downPos := m.active + m.cols
	for i := 0; i < len(m.inputs); i += 1 {
		if downPos-i >= 0 && downPos-i < len(m.inputs) && m.inputs[downPos-i] != nil {
			nextActive = downPos - i
			break
		}
		if downPos+i >= 0 && downPos+i < len(m.inputs) && m.inputs[downPos+i] != nil {
			nextActive = downPos + i
			break
		}
	}

	m.active = nextActive
	m.activateInput()
}
