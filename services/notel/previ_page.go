package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/NoelM/minigo"
)

func NewPrevisionPage(mntl *minigo.Minitel, communeMap map[string]string) *minigo.Page {
	previPage := minigo.NewPage("previsions", mntl, communeMap)

	var forecast OpenWeatherApiResponse
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

		body, err := getRequestBody(fmt.Sprintf(OWApiUrlFormat, commune.Latitude, commune.Longitude, OWApiKey))
		if err != nil {
			errorLog.Printf("unable to get forecasts: %s\n", err.Error())
			return sommaireId
		}
		defer body.Close()

		data := make([]byte, 100_000)
		n, _ := body.Read(data)

		if err := json.Unmarshal(data[:n], &forecast); err != nil {
			errorLog.Printf("unable to parse JSON: %s\n", err.Error())
			return sommaireId
		}

		printForecast(mntl, forecast, forecastId, commune)

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

func printForecast(mntl *minigo.Minitel, f OpenWeatherApiResponse, fId int, c Commune) {

	mntl.CursorOff()

	mntl.WriteAttributes(minigo.DoubleGrandeur, minigo.InversionFond)
	mntl.WriteStringLeft(2, c.NomCommune)
	mntl.WriteAttributes(minigo.GrandeurNormale, minigo.FondNormal)

	forecastTime, err := time.Parse("2006-01-02 15:04:05", date)
	if err != nil {
		warnLog.Printf("ignored entry %s: %s\n", date, err.Error())
	}

	mntl.WriteStringLeft(4, fmt.Sprintf("PREVISIONS LE %s %s A %s", weekdayIdToStringShort(forecastTime.Weekday()), forecastTime.Format("02/01"), forecastTime.Format("15:04")))
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
