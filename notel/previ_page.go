package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/logs"
	"github.com/NoelM/minigo/notel/utils"
)

func NewPrevisionPage(mntl *minigo.Minitel, communeMap map[string]string) *minigo.Page {
	previPage := minigo.NewPage("previsions", mntl, communeMap)

	var forecast OpenWeatherApiResponse
	var commune Commune

	// Setups the range of forecasts now -> now + 8 days
	now := time.Now()
	// The range of forecast goes from 00:00 to 21:00 UTC
	// If now is beyond 21:00, we skip to the next day
	if now.Hour() >= 21 {
		now = now.Add(4 * time.Hour)
	}
	firstForecastDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	forecastDate := firstForecastDate
	lastForecastDate := forecastDate.Add(5 * 24 * time.Hour) // 5 days of forecasts

	previPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.CursorOff()

		communeJSON, ok := initData["commune"]
		if !ok {
			logs.ErrorLog("no commune data\n")
			return sommaireId
		}

		if err := json.Unmarshal([]byte(communeJSON), &commune); err != nil {
			logs.ErrorLog("unable to parse the commune JSON: %s\n", err.Error())
			return sommaireId
		}

		OWApiKey := os.Getenv("OWAPIKEY")
		body, err := getRequestBody(fmt.Sprintf(OWApiUrlFormat, commune.Latitude, commune.Longitude, OWApiKey))
		if err != nil {
			logs.ErrorLog("unable to get forecasts: %s\n", err.Error())
			return sommaireId
		}
		defer body.Close()

		data, err := io.ReadAll(body)
		if err != nil {
			logs.ErrorLog("unable to get API response: %s\n", err.Error())
			return sommaireId
		}

		if err := json.Unmarshal(data, &forecast); err != nil {
			logs.ErrorLog("unable to parse JSON: %s\n", err.Error())
			return sommaireId
		}
		printForecast(mntl, forecast, forecastDate, commune)
		printPreviHelpers(mntl, forecastDate, firstForecastDate, lastForecastDate)

		return minigo.NoOp
	})

	previPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if forecastDate.Equal(lastForecastDate) {
			return nil, minigo.NoOp
		} else if forecastDate.After(lastForecastDate) {
			forecastDate = lastForecastDate
			return nil, minigo.NoOp
		}

		forecastDate = forecastDate.Add(24 * time.Hour)
		printForecast(mntl, forecast, forecastDate, commune)
		printPreviHelpers(mntl, forecastDate, firstForecastDate, lastForecastDate)

		return nil, minigo.NoOp
	})

	previPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if forecastDate.Equal(firstForecastDate) {
			return nil, minigo.NoOp
		} else if forecastDate.Before(firstForecastDate) {
			forecastDate = firstForecastDate
			return nil, minigo.NoOp
		}

		forecastDate = forecastDate.Add(-24 * time.Hour)
		printForecast(mntl, forecast, forecastDate, commune)
		printPreviHelpers(mntl, forecastDate, firstForecastDate, lastForecastDate)

		return nil, minigo.NoOp
	})

	previPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	return previPage
}

func printPreviHelpers(mntl *minigo.Minitel, forecastDate, firstForecastDate, lastForecastDate time.Time) {
	if forecastDate.After(firstForecastDate) {
		mntl.WriteHelperLeftAt(23, forecastDate.Add(-24*time.Hour).Format("02/01"), "RETOUR")
	}
	if forecastDate.Before(lastForecastDate) {
		mntl.WriteHelperRightAt(23, forecastDate.Add(24*time.Hour).Format("02/01"), "SUITE")
	}
	mntl.WriteHelperLeftAt(24, "Menu INFOMETEO", "SOMMAIRE")
}

func printForecast(mntl *minigo.Minitel, forecast OpenWeatherApiResponse, forecastDate time.Time, c Commune) {
	mntl.CleanScreen()
	location, _ := time.LoadLocation("Europe/Paris")

	// City name
	mntl.WriteAttributes(minigo.DoubleHauteur)
	if len(c.NomCommune) >= minigo.ColonnesSimple {
		mntl.WriteStringLeftAt(2, c.NomCommune[:minigo.ColonnesSimple-1])
	} else {
		mntl.WriteStringLeftAt(2, c.NomCommune)
	}
	mntl.WriteAttributes(minigo.GrandeurNormale)

	// Date of the forecast
	mntl.WriteStringLeftAt(3, fmt.Sprintf("%s %d %s",
		utils.WeekdayIdToString(forecastDate.Weekday()),
		forecastDate.Day(),
		utils.MonthIdToString(forecastDate.Month()),
	))

	maxTemp := 0.
	minTemp := 100.

	maxPress := 0
	minPress := 10000

	minWind := 100.
	maxWind := 0.

	// Print previsions
	lineId := 5
	for fDate := forecastDate; fDate.Before(forecastDate.Add(24 * time.Hour)); fDate = fDate.Add(3 * time.Hour) {
		for _, fct := range forecast.List {
			hourlyDate := time.Unix(fct.Dt, 0)
			if hourlyDate.After(fDate.Add(-time.Hour)) && hourlyDate.Before(fDate.Add(time.Hour)) {
				previsionString := fmt.Sprintf("%s> %2.f°C %s",
					fDate.In(location).Format("15:04"),
					fct.Main.Temp,
					weatherConditionCodeToString(fct.Weather[0].ID, fDate.In(location)))

				mntl.WriteStringLeftAt(lineId, previsionString)
				lineId += 2

				if fct.Main.Temp < minTemp {
					minTemp = fct.Main.Temp
				}
				if fct.Main.Temp > maxTemp {
					maxTemp = fct.Main.Temp
				}

				if fct.Wind.Speed < minWind {
					minWind = fct.Wind.Speed
				}
				if fct.Wind.Speed > maxWind {
					maxWind = fct.Wind.Speed
				}

				if fct.Main.Pressure < minPress {
					minPress = fct.Main.Pressure
				}
				if fct.Main.Pressure > maxPress {
					maxPress = fct.Main.Pressure
				}

			}
		}
	}

	mntl.WriteStringAtWithAttributes(5, 25, "Températures", minigo.InversionFond)
	mntl.WriteStringAt(6, 25, fmt.Sprintf("Min: %2.f°C", minTemp))
	mntl.WriteStringAt(7, 25, fmt.Sprintf("Max: %2.f°C", maxTemp))

	mntl.WriteStringAtWithAttributes(9, 25, fmt.Sprintf("Vent"), minigo.InversionFond)
	mntl.WriteStringAt(10, 25, fmt.Sprintf("Min: %2.f km/h", 3.6*minWind))
	mntl.WriteStringAt(11, 25, fmt.Sprintf("Max: %2.f km/h", 3.6*maxWind))

	mntl.WriteStringAtWithAttributes(13, 25, "Pression", minigo.InversionFond)
	mntl.WriteStringAt(14, 25, fmt.Sprintf("Min: %d hPa", minPress))
	mntl.WriteStringAt(15, 25, fmt.Sprintf("Max: %d hPa", maxPress))

	mntl.WriteStringAtWithAttributes(17, 25, "Ephéméride", minigo.InversionFond)
	mntl.WriteStringAt(18, 25, fmt.Sprintf("Lev.: %s",
		time.Unix(forecast.City.Sunrise, 0).In(location).Format("15:04")))
	mntl.WriteStringAt(19, 25, fmt.Sprintf("Cou.: %s",
		time.Unix(forecast.City.Sunset, 0).In(location).Format("15:04")))
}
