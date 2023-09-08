package minigo

type InputGroup struct {
	inputs             []*Input
	inputNames         []string
	active             int
	allowActiveCycling bool
}

func (ig *InputGroup) AppendInput(name string, m *Minitel, refX, refY int, width, height int, pre string, cursor bool) {
	ig.inputs = append(ig.inputs, NewInput(m, refX, refY, width, height, pre, cursor))
	ig.inputNames = append(ig.inputNames, name)
}

func (ig *InputGroup) ToMap() map[string]string {
	out := make(map[string]string)
	for i := 0; i < len(ig.inputs); i += 1 {
		out[ig.inputNames[i]] = string(ig.inputs[i].Value)
	}

	return out
}

func (ig *InputGroup) AppendKeyActive(key byte) {
	ig.inputs[ig.active].AppendKey(key)
}

func (ig *InputGroup) CorrectionActive() {
	ig.inputs[ig.active].Correction()
}

func (ig *InputGroup) RepetitionActive() {
	ig.inputs[ig.active].Repetition()
}

func (ig *InputGroup) RepetitionAll() {
	for i := 0; i < len(ig.inputs); i += 1 {
		ig.inputs[i].Repetition()
	}
}

func (ig *InputGroup) ClearScreenActive() {
	ig.inputs[ig.active].ClearScreen()
}

func (ig *InputGroup) ClearScreenAll() {
	for i := 0; i < len(ig.inputs); i += 1 {
		ig.inputs[i].ClearScreen()
	}
}

func (ig *InputGroup) ClearActive() {
	ig.inputs[ig.active].Clear()
}

func (ig *InputGroup) ClearAll() {
	for i := 0; i < len(ig.inputs); i += 1 {
		ig.inputs[i].Clear()
	}
}

func (ig *InputGroup) activateActive() {
	ig.inputs[ig.active].Activate()
}

func (ig *InputGroup) ActivateFirst() {
	ig.active = 0
	ig.activateActive()
}

func (ig *InputGroup) ActivateNext() {
	ig.active += 1
	if ig.active >= len(ig.inputs) {
		if ig.allowActiveCycling {
			ig.active = 0
		} else {
			ig.active = len(ig.inputs)
		}
	}
	ig.activateActive()
}

func (ig *InputGroup) ActivatePrev() {
	ig.active -= 1
	if ig.active < 0 {
		if ig.allowActiveCycling {
			ig.active = len(ig.inputs) - 1
		} else {
			ig.active = 0
		}
	}
	ig.activateActive()
}
