package minigo

type Page struct {
	m *Minitel

	initFunc func()
}

func (p *Page) init() {
	p.initFunc()
}

func (p *Page) Run() ([]byte, int) {

	p.init()

	for {
		select {
		case key := <-m.RecvKey:
			if key == minigo.Envoi {
				if len(nickInput.Value) == 0 {
					warnLog.Println("Empty nick input")
					continue
				}
				m.Reset()

				infoLog.Printf("Logged as: %s\n", nickInput.Value)
				return nickInput.Value, noopId

			} else if key == minigo.Correction {
				nickInput.Correction()

			} else if key == minigo.Sommaire {
				return nil, sommaireId

			} else if minigo.IsUintAValidChar(key) {
				nickInput.AppendKey(byte(key))

			} else {
				errorLog.Printf("Not supported key: %d\n", key)
			}

		case <-m.Quit:
			warnLog.Println("Quitting log page")
			return nil, quitId

		default:
			continue
		}
	}
}
