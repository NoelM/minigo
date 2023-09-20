package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/NoelM/minigo"
)

const APIForecastFormat = "http://www.infoclimat.fr/public-api/gfs/json?_ll=%.5f,%f.5&_auth=U0kEEwZ4U3FVeAE2VyELIlgwBDFdKwEmB3sFZgBoUi8CYANhB2IAZFA4UC1VegAxUXwPaA0zAjxQNVc3AXNTL1MzBGAGZFM1VT0BYVdvCyBYdAR5XWMBJgd7BWsAb1IvAmADYAdkAHxQNlAsVWcANlFkD3ANLQI7UDVXOQFoUzNTNARlBm1TMVU7AXxXeAs5WG8EMl02AWgHNQVmAGRSNwI0AzYHNwBrUD1QLFVjADJRYA9tDTICP1A2VzIBc1MvU0kEEwZ4U3FVeAE2VyELIlg%%2BBDpdNg%%3D%%3D&_c=940e429e25a778ab4196831fbc0d51b8"

func NewPrevisionPage(mntl *minigo.Minitel, communeMap map[string]string) *minigo.Page {
	previPage := minigo.NewPage("previsions", mntl, communeMap)

	var forecast APIForecastReply
	var forecastSort map[int]string
	var commune Commune

	forecastId := 0

	previPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		communeJSON, ok := initData["commune"]
		if !ok {
			errorLog.Println("no commune data for prevision")
			return sommaireId
		}

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

		forecastSort = make(map[int]string)
		sortForecasts(&forecast, forecastSort)
		printForecast(mntl, forecast.Forecasts[forecastSort[forecastId]], forecastSort[forecastId], &commune)

		return minigo.NoOp
	})

	previPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		infoLog.Println("request suite")
		forecastId += 1
		if forecastId >= len(forecastSort) {
			forecastId = len(forecastSort) - 1
			return nil, minigo.NoOp
		}
		printForecast(mntl, forecast.Forecasts[forecastSort[forecastId]], forecastSort[forecastId], &commune)
		return nil, minigo.NoOp
	})

	previPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		infoLog.Println("request retour")
		forecastId -= 1
		if forecastId < 0 {
			forecastId = 0
			return nil, minigo.NoOp
		}
		printForecast(mntl, forecast.Forecasts[forecastSort[forecastId]], forecastSort[forecastId], &commune)
		return nil, minigo.NoOp
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

	for k := range f.Forecasts {
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

	mntl.CursorOff()

	mntl.WriteAttributes(minigo.DoubleGrandeur, minigo.InversionFond)
	mntl.WriteStringLeft(2, c.NomCommune)
	mntl.WriteAttributes(minigo.GrandeurNormale, minigo.FondNormal)

	forecastTime, err := time.Parse("2006-01-02 15:04:05", date)
	if err != nil {
		warnLog.Printf("ignored entry %s: %s\n", date, err.Error())
	}

	mntl.WriteStringLeft(4, fmt.Sprintf("PREVISIONS LE %s %s A %s", weekDayIdToString(forecastTime.Weekday()), forecastTime.Format("02/01"), forecastTime.Format("15:04")))
	mntl.WriteStringLeft(5, "DONNEES: INFO-CLIMAT")

	mntl.CleanScreenFromXY(1, 7)

	mntl.WriteStringAtWithAttributes(1, 7, nebulositeToString(f.Nebulosite.Totale), minigo.InversionFond)

	mntl.WriteStringLeft(9, fmt.Sprintf("TEMP: %.0f C", f.Temperature.TwoM-275.))
	if f.VentRafales.One0M > 10 && f.VentRafales.One0M > f.VentMoyen.One0M {
		mntl.WriteStringLeft(11, fmt.Sprintf("VENT: %.0f km/h - RAFALES: %.0f km/h", f.VentMoyen.One0M, f.VentRafales.One0M))
	} else {
		mntl.WriteStringLeft(11, fmt.Sprintf("VENT: %.0f km/h", f.VentMoyen.One0M*3.6))
	}
	mntl.WriteStringLeft(12, fmt.Sprintf("DIR:  %s", windDirToString(f.VentDirection.One0M)))
	mntl.WriteStringLeft(14, fmt.Sprintf("PLUIE: %.0f mm", f.Pluie))

	mntl.WriteHelperLeft(23, "PREC.", "RETOUR")
	mntl.WriteHelperRight(23, "SUIV.", "SUITE")
	mntl.WriteHelperLeft(24, "CHOIX CODE POSTAL", "SOMMAIRE")
}

func nebulositeToString(n float64) string {
	octas := 8. * n / 100.

	if octas < 1 {
		return "CIEL CLAIR"
	} else if octas >= 1 && octas < 2 {
		return "PEU NUAGEUX"
	} else if octas >= 2 && octas < 5 {
		return "NUAGEUX"
	} else if octas >= 5 && octas < 7 {
		return "TRES NUAGEUX"
	} else if octas >= 7 {
		return "CIEL COUVERT"
	}
	return ""
}

func weekDayIdToString(i time.Weekday) string {
	switch i {
	case time.Sunday:
		return "Dim."
	case time.Monday:
		return "Lun."
	case time.Tuesday:
		return "Mar."
	case time.Wednesday:
		return "Mer."
	case time.Thursday:
		return "Jeu."
	case time.Friday:
		return "Ven."
	case time.Saturday:
		return "Sam."
	}
	return ""
}

func monthIdToString(i time.Month) string {
	switch i {
	case time.January:
		return "Janvier"
	case time.February:
		return "Février"
	case time.March:
		return "Mars"
	case time.April:
		return "Avril"
	case time.May:
		return "Mai"
	case time.June:
		return "Juin"
	case time.July:
		return "Juillet"
	case time.August:
		return "Août"
	case time.September:
		return "Septembre"
	case time.October:
		return "Octobre"
	case time.November:
		return "Novembre"
	case time.December:
		return "Décembre"
	}
	return ""
}
