package main

import "github.com/NoelM/minigo"

func ServiceMeteo(m *minigo.Minitel) int {
HOME:
	out, serviceId := NewCodePostalPage(m).Run()
	if serviceId == minigo.SuiteOp {
		goto OBS
	} else if serviceId != minigo.NoOp && serviceId != minigo.QuitOp {
		return serviceId
	}

	goto PREVI

OBS:
	out, serviceId = NewObservationsPage(m).Run()
	if serviceId == sommaireId {
		goto HOME
	} else if serviceId != minigo.NoOp && serviceId != minigo.QuitOp {
		return serviceId
	}

PREVI:
	out, serviceId = NewCommunesPage(m, out).Run()
	if serviceId == sommaireId {
		goto HOME
	} else if serviceId != minigo.NoOp && serviceId != minigo.QuitOp {
		return serviceId
	}

	_, serviceId = NewPrevisionPage(m, out).Run()
	if serviceId == sommaireId {
		goto HOME
	} else if serviceId != minigo.NoOp && serviceId != minigo.QuitOp {
		return serviceId
	}

	return sommaireId
}
