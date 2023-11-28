package minigo

const (
	NoOp = iota - 50
	DisconnectOp
	QuitOp
	EnvoiOp
	SommaireOp
	AnnulationOp
	RetourOp
	SuiteOp
	RepetitionOp
	GuideOp
	CorrectionOp
)

type InitFunc func(mntl *Minitel, inputs *Form, initData map[string]string) int
type KeyboardFunc func(mntl *Minitel, inputs *Form, key rune)
type NavigationFunc func(mntl *Minitel, inputs *Form) (map[string]string, int)
type ConnexionFinFunc func(mntl *Minitel) int

type Page struct {
	mntl     *Minitel
	name     string
	initData map[string]string
	form     *Form

	initFunc         InitFunc
	charFunc         KeyboardFunc
	connexionFinFunc ConnexionFinFunc
	envoiFunc        NavigationFunc
	sommaireFunc     NavigationFunc
	annulationFunc   NavigationFunc
	retourFunc       NavigationFunc
	repetitionFunc   NavigationFunc
	guideFunc        NavigationFunc
	correctionFunc   NavigationFunc
	suiteFunc        NavigationFunc
	hautFunc         NavigationFunc
	basFunc          NavigationFunc
	droiteFunc       NavigationFunc
	gaucheFunc       NavigationFunc
}

func NewPage(name string, mntl *Minitel, initData map[string]string) *Page {
	return &Page{
		mntl:             mntl,
		name:             name,
		initData:         initData,
		initFunc:         func(mntl *Minitel, inputs *Form, initData map[string]string) int { return NoOp },
		charFunc:         func(mntl *Minitel, inputs *Form, key rune) {},
		connexionFinFunc: defaultConnexionFinHandlerFunc,
		envoiFunc:        defaultNavigationHandlerFunc,
		sommaireFunc:     defaultNavigationHandlerFunc,
		annulationFunc:   defaultNavigationHandlerFunc,
		retourFunc:       defaultNavigationHandlerFunc,
		repetitionFunc:   defaultNavigationHandlerFunc,
		guideFunc:        defaultNavigationHandlerFunc,
		correctionFunc:   defaultNavigationHandlerFunc,
		suiteFunc:        defaultNavigationHandlerFunc,
		hautFunc:         defaultNavigationHandlerFunc,
		basFunc:          defaultNavigationHandlerFunc,
		droiteFunc:       defaultNavigationHandlerFunc,
		gaucheFunc:       defaultNavigationHandlerFunc,
	}
}

func defaultConnexionFinHandlerFunc(mntl *Minitel) int {
	mntl.CleanScreen()
	mntl.WriteStringLeft(1, "→ Déconnexion demandée, à bientôt !")
	return DisconnectOp
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

func (p *Page) SetHautFunc(f NavigationFunc) {
	p.hautFunc = f
}

func (p *Page) SetBasFunc(f NavigationFunc) {
	p.basFunc = f
}

func (p *Page) SetDroiteFunc(f NavigationFunc) {
	p.droiteFunc = f
}

func (p *Page) SetGaucheFunc(f NavigationFunc) {
	p.gaucheFunc = f
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
		key := <-p.mntl.In

		switch key {
		case Envoi:
			if out, op := p.envoiFunc(p.mntl, p.form); op != NoOp {
				infoLog.Printf("page: key Envoi: quit %s page, with op=%d\n", p.name, op)
				return out, op
			}

		case Sommaire:
			if out, op := p.sommaireFunc(p.mntl, p.form); op != NoOp {
				infoLog.Printf("page: key Sommaire: quit %s page, with op=%d\n", p.name, op)
				return out, op
			}

		case Annulation:
			if out, op := p.annulationFunc(p.mntl, p.form); op != NoOp {
				infoLog.Printf("page: key Annulation: quit %s page, with op=%d\n", p.name, op)
				return out, op
			}

		case Retour:
			if out, op := p.retourFunc(p.mntl, p.form); op != NoOp {
				infoLog.Printf("page: key Retour: quit %s page, with op=%d\n", p.name, op)
				return out, op
			}

		case Repetition:
			if out, op := p.repetitionFunc(p.mntl, p.form); op != NoOp {
				infoLog.Printf("page: key Repetition: quit %s page, with op=%d\n", p.name, op)
				return out, op
			}

		case Guide:
			if out, op := p.guideFunc(p.mntl, p.form); op != NoOp {
				infoLog.Printf("page: key Guide: quit %s page, with op=%d\n", p.name, op)
				return out, op
			}

		case Correction:
			if out, op := p.correctionFunc(p.mntl, p.form); op != NoOp {
				infoLog.Printf("page: key Correction: quit %s page, with op=%d\n", p.name, op)
				return out, op
			}

		case Suite:
			if out, op := p.suiteFunc(p.mntl, p.form); op != NoOp {
				infoLog.Printf("page: key Suite: quit %s page, with op=%d\n", p.name, op)
				return out, op
			}

		case ToucheFlecheHaut:
			if out, op := p.hautFunc(p.mntl, p.form); op != NoOp {
				infoLog.Printf("page: key Haut: quit %s page, with op=%d\n", p.name, op)
				return out, op
			}

		case ToucheFlecheBas:
			if out, op := p.basFunc(p.mntl, p.form); op != NoOp {
				infoLog.Printf("page: key Bas: quit %s page, with op=%d\n", p.name, op)
				return out, op
			}

		case ToucheFlecheDroite:
			if out, op := p.droiteFunc(p.mntl, p.form); op != NoOp {
				infoLog.Printf("page: key Droite: quit %s page, with op=%d\n", p.name, op)
				return out, op
			}

		case ToucheFlecheGauche:
			if out, op := p.gaucheFunc(p.mntl, p.form); op != NoOp {
				infoLog.Printf("page: key Gauche: quit %s page, with op=%d\n", p.name, op)
				return out, op
			}

		case ConnexionFin:
			if op := p.connexionFinFunc(p.mntl); op != NoOp {
				infoLog.Printf("page: key ConnexionFin: quit page %s, with op=%d\n", p.name, op)
				return nil, op
			}

		case PCE:
			p.form = &Form{}
			if op := p.initFunc(p.mntl, p.form, p.initData); op != NoOp {
				return nil, op
			}

		default:
			if ValidRune(key) {
				p.charFunc(p.mntl, p.form, key)
			} else {
				errorLog.Printf("page: invalid rune=%x\n", key)
			}
		}
	}
}
