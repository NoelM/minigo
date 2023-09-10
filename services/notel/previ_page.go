package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/NoelM/minigo"
)

const APIForecastFormat = "http://www.infoclimat.fr/public-api/gfs/json?_ll=%.5f,%f.5&_auth=U0kEEwZ4U3FVeAE2VyELIlgwBDFdKwEmB3sFZgBoUi8CYANhB2IAZFA4UC1VegAxUXwPaA0zAjxQNVc3AXNTL1MzBGAGZFM1VT0BYVdvCyBYdAR5XWMBJgd7BWsAb1IvAmADYAdkAHxQNlAsVWcANlFkD3ANLQI7UDVXOQFoUzNTNARlBm1TMVU7AXxXeAs5WG8EMl02AWgHNQVmAGRSNwI0AzYHNwBrUD1QLFVjADJRYA9tDTICP1A2VzIBc1MvU0kEEwZ4U3FVeAE2VyELIlg%%2BBDpdNg%%3D%%3D&_c=940e429e25a778ab4196831fbc0d51b8"

func NewPrevisionPage(mntl *minigo.Minitel, commune map[string]string) *minigo.Page {
	previPage := minigo.NewPage("previsions", mntl, commune)

	var forecast APIForecastReply
	var forecastSort map[int]string

	forecastId := 0

	previPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		communeJSON, ok := initData["commune"]
		if !ok {
			errorLog.Println("no commune data for prevision")
			return sommaireId
		}

		var commune Commune
		if err := json.Unmarshal([]byte(communeJSON), &commune); err != nil {
			errorLog.Printf("unable to parse the commune JSON: %s\n", err.Error())
			return sommaireId
		}

		body, err := getRequestBody(fmt.Sprintf(APIForecastFormat, commune.Latitude, commune.Longitude))
		if err != nil {
			errorLog.Printf("unable to get forecasts: %s\n", err.Error())
			return sommaireId
		}
		defer body.Close()

		data := make([]byte, 100_000)
		n, _ := body.Read(data)

		if err := forecast.UnmarshalJSON(data[:n]); err != nil {
			errorLog.Printf("unable to parse JSON: %s\n", err.Error())
			return sommaireId
		}

		sortForecasts(&forecast, forecastSort)
		printForecast(mntl, forecast.Forecasts[forecastSort[forecastId]], forecastSort[forecastId], &commune)

		return minigo.NoOp
	})

	previPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	return previPage
}

type SortableEpochs struct {
	items []int64
}

func (s *SortableEpochs) Len() int           { return len(s.items) }
func (s *SortableEpochs) Less(i, j int) bool { return s.items[i] < s.items[j] }
func (s *SortableEpochs) Swap(i, j int)      { tmp := s.items[i]; s.items[i] = s.items[j]; s.items[j] = tmp }

func sortForecasts(f *APIForecastReply, order map[int]string) {
	epochToText := make(map[int64]string)
	var epochs SortableEpochs

	for k, _ := range f.Forecasts {
		if len(k) > 0 && k[0] == '2' {
			forecastTime, err := time.Parse("2006-01-02 15:04:05", k)
			if err != nil {
				warnLog.Printf("ignored entry %s: %s\n", k, err.Error())
			}

			epochToText[forecastTime.Unix()] = k
			epochs.items = append(epochs.items, forecastTime.Unix())
		}
	}

	sort.Sort(&epochs)

	for id, epoch := range epochs.items {
		order[id] = epochToText[epoch]
	}
}

func printForecast(mntl *minigo.Minitel, f Forecast, date string, c *Commune) {

	mntl.WriteAttributes(minigo.DoubleGrandeur, minigo.InversionFond)
	mntl.WriteStringXY(1, 2, c.NomCommune)
	mntl.WriteAttributes(minigo.GrandeurNormale, minigo.FondNormal)

	forecastTime, err := time.Parse("2006-01-02 15:04:05", date)
	if err != nil {
		warnLog.Printf("ignored entry %s: %s\n", date, err.Error())
	}

	mntl.WriteStringXY(1, 4, fmt.Sprintf("PREVISIONS POUR LE %s", forecastTime.Format("01/02/06 15:04")))

	mntl.WriteStringXY(1, 6, fmt.Sprintf("TEMP: %.0f C", f.Temperature.TwoM-275.))
	mntl.WriteStringXY(1, 7, fmt.Sprintf("VENT: %.0f km/h", f.VentMoyen.One0M*3.6))
}
