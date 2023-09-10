package main

import (
	"testing"
)

func TestGetCodePostal(t *testing.T) {
	loadCommuneDatabase()
	communes := getCommunesFromCodePostal("07100")
	if communes == nil {
		t.Fatal("unable to fetch API")
	}
	t.Log(communes)
}
