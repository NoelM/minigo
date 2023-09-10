package main

import "github.com/NoelM/minigo"

func ServiceMeteo(m *minigo.Minitel) int {
	out, serviceId := NewCodePostalPage(m).Run()
	if serviceId != minigo.NoOp && serviceId != minigo.QuitOp {
		return serviceId
	}

	_, serviceId = NewCommunesPage(m, out).Run()
	if serviceId != minigo.NoOp && serviceId != minigo.QuitOp {
		return serviceId
	}

	return sommaireId
}
