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

func (f *Form) InitAll() {
	for _, i := range f.inputs {
		i.Init()
	}
	f.ActivateFirst()
}

func (f *Form) AppendKeyActive(key rune) {
	f.inputs[f.active].AppendKey(key)
}

func (f *Form) CorrectionActive() {
	f.inputs[f.active].Correction()
}

func (f *Form) UnHideActive() {
	f.inputs[f.active].UnHide()
}

func (f *Form) UnHideAll() {
	for i := 0; i < len(f.inputs); i += 1 {
		f.inputs[i].UnHide()
	}
	f.activateInput()
}

func (f *Form) HideActive() {
	f.inputs[f.active].Hide()
}

func (f *Form) HideAll() {
	for i := 0; i < len(f.inputs); i += 1 {
		f.inputs[i].Hide()
	}
}

func (f *Form) ResetActive() {
	f.inputs[f.active].Reset()
}

func (f *Form) ResetAll() {
	for i := 0; i < len(f.inputs); i += 1 {
		f.inputs[i].Reset()
	}
	f.ActivateFirst()
}

func (f *Form) activateInput() {
	f.inputs[f.active].Activate()
}

func (f *Form) ActivateFirst() {
	f.active = 0
	f.activateInput()
}

func (f *Form) ActivateNext() {
	f.active += 1
	if f.active >= len(f.inputs) {
		if f.allowActiveCycling {
			f.active = 0
		} else {
			f.active = len(f.inputs) - 1
		}
	}
	f.activateInput()
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
	f.activateInput()
}
