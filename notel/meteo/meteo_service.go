package meteo

import (
	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

func ServiceMeteo(m *minigo.Minitel, communeDB *databases.CommuneDatabase) int {
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
	if serviceId == minigo.SommaireOp {
		goto HOME
	} else if serviceId != minigo.NoOp && serviceId != minigo.QuitOp {
		return serviceId
	}

PREVI:
	out, serviceId = NewCommunesPage(m, communeDB, out).Run()
	if serviceId == minigo.SommaireOp {
		goto HOME
	} else if serviceId != minigo.NoOp && serviceId != minigo.QuitOp {
		return serviceId
	}

	_, serviceId = NewPrevisionPage(m, out).Run()
	if serviceId == minigo.SommaireOp {
		goto HOME
	} else if serviceId != minigo.NoOp && serviceId != minigo.QuitOp {
		return serviceId
	}

	return minigo.SommaireOp
}
