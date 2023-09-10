package main

const APICodePostauxFormat = "https://apicarto.ign.fr/api/codes-postaux/communes/%s"

type Commune struct {
	CodeCommune string `json:"codeCommune"`
	CodePostal  string `json:"CodePostal"`
	NomCommune  string `json:"nomCommune"`
}
