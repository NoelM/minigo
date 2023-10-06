package minigo

const (
	NoOp = iota - 10
	DisconnectOp
	QuitPageOp
	SuiteOp
)

type InitFunc func(mntl *Minitel, inputs *Form, initData map[string]string) int
type KeyboardFunc func(mntl *Minitel, inputs *Form, key uint)
type InChanFunc func(mntl *Minitel, inputs *Form, message string)
type NavigationFunc func(mntl *Minitel, inputs *Form) (map[string]string, int)
type ConnexionFinFunc func(mntl *Minitel) int

type Page struct {
	InChan  chan string
	OutChan chan string

	mntl     *Minitel
	name     string
	initData map[string]string
	form     *Form

	initFunc         InitFunc
	charFunc         KeyboardFunc
	inChanFunc       InChanFunc
	connexionFinFunc ConnexionFinFunc
	envoiFunc        NavigationFunc
	sommaireFunc     NavigationFunc
	annulationFunc   NavigationFunc
	retourFunc       NavigationFunc
	repetitionFunc   NavigationFunc
	guideFunc        NavigationFunc
	correctionFunc   NavigationFunc
	suiteFunc        NavigationFunc
}

func NewPage(name string, mntl *Minitel, initData map[string]string) *Page {
	return &Page{
		mntl:             mntl,
		name:             name,
		initData:         initData,
		initFunc:         func(mntl *Minitel, inputs *Form, initData map[string]string) int { return NoOp },
		charFunc:         func(mntl *Minitel, inputs *Form, key uint) {},
		inChanFunc:       func(mntl *Minitel, inputs *Form, message string) {},
		connexionFinFunc: func(mntl *Minitel) int { mntl.Disconnect(); return DisconnectOp },
		envoiFunc:        defaultNavigationHandlerFunc,
		sommaireFunc:     defaultNavigationHandlerFunc,
		annulationFunc:   defaultNavigationHandlerFunc,
		retourFunc:       defaultNavigationHandlerFunc,
		repetitionFunc:   defaultNavigationHandlerFunc,
		guideFunc:        defaultNavigationHandlerFunc,
		correctionFunc:   defaultNavigationHandlerFunc,
		suiteFunc:        defaultNavigationHandlerFunc,
	}
}

func defaultNavigationHandlerFunc(mntl *Minitel, input *Form) (map[string]string, int) {
	return nil, NoOp
}

func (p *Page) SetInitFunc(f InitFunc) {
	p.initFunc = f
}

func (p *Page) SetCharFunc(f KeyboardFunc) {
	p.charFunc = f
}

func (p *Page) SetEnvoiFunc(f NavigationFunc) {
	p.envoiFunc = f
}

func (p *Page) SetSommaireFunc(f NavigationFunc) {
	p.sommaireFunc = f
}

func (p *Page) SetAnnulationFunc(f NavigationFunc) {
	p.annulationFunc = f
}

func (p *Page) SetRetourFunc(f NavigationFunc) {
	p.retourFunc = f
}

func (p *Page) SetRepetitionFunc(f NavigationFunc) {
	p.repetitionFunc = f
}

func (p *Page) SetGuideFunc(f NavigationFunc) {
	p.guideFunc = f
}

func (p *Page) SetCorrectionFunc(f NavigationFunc) {
	p.correctionFunc = f
}

func (p *Page) SetSuiteFunc(f NavigationFunc) {
	p.suiteFunc = f
}

func (p *Page) SetInChanFunc(f InChanFunc) {
	p.inChanFunc = f
}

func (p *Page) SetConnexionFinFunc(f ConnexionFinFunc) {
	p.connexionFinFunc = f
}

func (p *Page) Run() (map[string]string, int) {

	p.form = &Form{}
	if op := p.initFunc(p.mntl, p.form, p.initData); op != NoOp {
		return nil, op
	}

	for {
		select {
		case msg := <-p.InChan:
			p.inChanFunc(p.mntl, p.form, msg)

		case key := <-p.mntl.RecvKey:
			switch key {
			case Envoi:
				if out, op := p.envoiFunc(p.mntl, p.form); op != NoOp {
					infoLog.Printf("key Envoi: quit page %s page, with op=%d\n", p.name, op)
					return out, op
				}

			case Sommaire:
				if out, op := p.sommaireFunc(p.mntl, p.form); op != NoOp {
					infoLog.Printf("key Sommaire: quit page %s page, with op=%d\n", p.name, op)
					return out, op
				}

			case Annulation:
				if out, op := p.annulationFunc(p.mntl, p.form); op != NoOp {
					infoLog.Printf("key Annulation: quit page %s page, with op=%d\n", p.name, op)
					return out, op
				}

			case Retour:
				if out, op := p.retourFunc(p.mntl, p.form); op != NoOp {
					infoLog.Printf("key Retour: quit page %s page, with op=%d\n", p.name, op)
					return out, op
				}

			case Repetition:
				if out, op := p.repetitionFunc(p.mntl, p.form); op != NoOp {
					infoLog.Printf("key Repetition: quit page %s page, with op=%d\n", p.name, op)
					return out, op
				}

			case Guide:
				if out, op := p.guideFunc(p.mntl, p.form); op != NoOp {
					infoLog.Printf("key Guide: quit page %s page, with op=%d\n", p.name, op)
					return out, op
				}

			case Correction:
				if out, op := p.correctionFunc(p.mntl, p.form); op != NoOp {
					infoLog.Printf("key Correction: quit page %s page, with op=%d\n", p.name, op)
					return out, op
				}

			case Suite:
				if out, op := p.suiteFunc(p.mntl, p.form); op != NoOp {
					infoLog.Printf("key Suite: quit page %s page, with op=%d\n", p.name, op)
					return out, op
				}

			case ConnexionFin:
				if op := p.connexionFinFunc(p.mntl); op != NoOp {
					infoLog.Printf("key ConnexionFin: disconnect %s page\n", p.name)
					return nil, op
				}

			default:
				if IsUintAValidChar(key) {
					p.charFunc(p.mntl, p.form, key)

				} else {
					errorLog.Printf("not supported key: %d\n", key)
				}
			}
		}
	}
}
