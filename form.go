package minigo

type Form struct {
	inputs             []*Input
	inputNames         []string
	active             int
	allowActiveCycling bool
}

func (f *Form) AppendInput(name string, input *Input) {
	f.inputs = append(f.inputs, input)
	f.inputNames = append(f.inputNames, name)
}

func (f *Form) ToMap() map[string]string {
	out := make(map[string]string)
	for i := 0; i < len(f.inputs); i += 1 {
		out[f.inputNames[i]] = string(f.inputs[i].Value)
	}

	return out
}

func (f *Form) ValueActive() string {
	return string(f.inputs[f.active].Value)
}

func (f *Form) AppendKeyActive(key rune) {
	f.inputs[f.active].AppendKey(key)
}

func (f *Form) CorrectionActive() {
	f.inputs[f.active].Correction()
}

func (f *Form) RepetitionActive() {
	f.inputs[f.active].Repetition()
}

func (f *Form) RepetitionAll() {
	for i := 0; i < len(f.inputs); i += 1 {
		f.inputs[i].Repetition()
	}
}

func (f *Form) ClearScreenActive() {
	f.inputs[f.active].ClearScreen()
}

func (f *Form) ClearScreenAll() {
	for i := 0; i < len(f.inputs); i += 1 {
		f.inputs[i].ClearScreen()
	}
}

func (f *Form) ClearActive() {
	f.inputs[f.active].Clear()
}

func (f *Form) ClearAll() {
	for i := 0; i < len(f.inputs); i += 1 {
		f.inputs[i].Clear()
	}
}

func (f *Form) activateActive() {
	f.inputs[f.active].Activate()
}

func (f *Form) ActivateFirst() {
	f.active = 0
	f.activateActive()
}

func (f *Form) ActivateNext() {
	f.active += 1
	if f.active >= len(f.inputs) {
		if f.allowActiveCycling {
			f.active = 0
		} else {
			f.active = len(f.inputs)
		}
	}
	f.activateActive()
}

func (f *Form) ActivatePrev() {
	f.active -= 1
	if f.active < 0 {
		if f.allowActiveCycling {
			f.active = len(f.inputs) - 1
		} else {
			f.active = 0
		}
	}
	f.activateActive()
}
