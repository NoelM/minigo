package main

import (
	"testing"
)

func TestGetCodePostal(t *testing.T) {
	cdb := CommuneDatabase{}
	cdb.LoadCommuneDatabase("/media/core/communes-departement-region.csv")
	communes := cdb.GetCommunesFromCodePostal("07100")
	if communes == nil {
		t.Fatal("unable to fetch API")
	}
	t.Log(communes)
}

func TestInfoClimatForecastParse(t *testing.T) {
	const URL = "http://www.infoclimat.fr/public-api/gfs/json?_ll=45.16667,5.71667&_auth=U0kEEwZ4U3FVeAE2VyELIlgwBDFdKwEmB3sFZgBoUi8CYANhB2IAZFA4UC1VegAxUXwPaA0zAjxQNVc3AXNTL1MzBGAGZFM1VT0BYVdvCyBYdAR5XWMBJgd7BWsAb1IvAmADYAdkAHxQNlAsVWcANlFkD3ANLQI7UDVXOQFoUzNTNARlBm1TMVU7AXxXeAs5WG8EMl02AWgHNQVmAGRSNwI0AzYHNwBrUD1QLFVjADJRYA9tDTICP1A2VzIBc1MvU0kEEwZ4U3FVeAE2VyELIlg%2BBDpdNg%3D%3D&_c=940e429e25a778ab4196831fbc0d51b8"

	body, err := getRequestBody(URL)
	if err != nil {
		t.Fatalf("cannot get body: %s\n", err.Error())
	}
	defer body.Close()

	data := make([]byte, 100_000)
	n, _ := body.Read(data)

	var forecast APIForecastReply
	if err := forecast.UnmarshalJSON(data[:n]); err != nil {
		t.Fatalf("unable to parse JSON: %s\n", err.Error())
	}

	t.Log(forecast.Forecasts)
}
