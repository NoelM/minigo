package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/NoelM/minigo"
)

func NewPrevisionPage(mntl *minigo.Minitel, communeMap map[string]string) *minigo.Page {
	previPage := minigo.NewPage("previsions", mntl, communeMap)

	var forecast OpenWeatherApiResponse
	var commune Commune

	// Setups the range of forecasts now -> now + 8 days
	now := time.Now()
	firstForecastDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	forecastDate := firstForecastDate
	lastForecastDate := forecastDate.Add(5 * 24 * time.Hour) // 5 days of forecasts

	previPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.CursorOff()

		communeJSON, ok := initData["commune"]
		if !ok {
			errorLog.Println("no commune data")
			return sommaireId
		}

		if err := json.Unmarshal([]byte(communeJSON), &commune); err != nil {
			errorLog.Printf("unable to parse the commune JSON: %s\n", err.Error())
			return sommaireId
		}

		OWApiKey := os.Getenv("OWAPIKEY")
		body, err := getRequestBody(fmt.Sprintf(OWApiUrlFormat, commune.Latitude, commune.Longitude, OWApiKey))
		if err != nil {
			errorLog.Printf("unable to get forecasts: %s\n", err.Error())
			return sommaireId
		}
		defer body.Close()

		data, err := io.ReadAll(body)
		if err != nil {
			errorLog.Printf("unable to get API response: %s\n", err.Error())
			return sommaireId
		}

		if err := json.Unmarshal(data, &forecast); err != nil {
			errorLog.Printf("unable to parse JSON: %s\n", err.Error())
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
		if hasForecast := printForecast(mntl, forecast, forecastDate, commune); !hasForecast {
			forecastDate = forecastDate.Add(24 * time.Hour)
			printForecast(mntl, forecast, forecastDate, commune)
		}
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
		if hasForecast := printForecast(mntl, forecast, forecastDate, commune); !hasForecast {
			forecastDate = forecastDate.Add(24 * time.Hour)
			printForecast(mntl, forecast, forecastDate, commune)
		}
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
		mntl.WriteHelperLeft(23, forecastDate.Add(-24*time.Hour).Format("02/01"), "RETOUR")
	}
	if forecastDate.Before(lastForecastDate) {
		mntl.WriteHelperRight(23, forecastDate.Add(24*time.Hour).Format("02/01"), "SUITE")
	}
	mntl.WriteHelperLeft(24, "Menu INFOMETEO", "SOMMAIRE")
}

func printForecast(mntl *minigo.Minitel, forecast OpenWeatherApiResponse, forecastDate time.Time, c Commune) (hasForecast bool) {
	mntl.CleanScreen()
	location, _ := time.LoadLocation("Europe/Paris")

	// City name
	mntl.WriteAttributes(minigo.DoubleHauteur)
	if len(c.NomCommune) >= minigo.ColonnesSimple {
		mntl.WriteStringLeft(2, c.NomCommune[:minigo.ColonnesSimple-1])
	} else {
		mntl.WriteStringLeft(2, c.NomCommune)
	}
	mntl.WriteAttributes(minigo.GrandeurNormale)

	// Date of the forecast
	mntl.WriteStringLeft(3, fmt.Sprintf("%s %d %s",
		weekdayIdToString(forecastDate.Weekday()),
		forecastDate.Day(),
		monthIdToString(forecastDate.Month()),
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

				mntl.WriteStringLeft(lineId, previsionString)
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

		// Has no forecast
		if lineId == 5 {
			return
		}
	}

	mntl.WriteStringAtWithAttributes(25, 5, fmt.Sprintf("Températures"), minigo.InversionFond)
	mntl.WriteStringAt(25, 6, fmt.Sprintf("Min: %2.f°C", minTemp))
	mntl.WriteStringAt(25, 7, fmt.Sprintf("Max: %2.f°C", maxTemp))

	mntl.WriteStringAtWithAttributes(25, 9, fmt.Sprintf("Vent"), minigo.InversionFond)
	mntl.WriteStringAt(25, 10, fmt.Sprintf("Min: %2.f km/h", minWind))
	mntl.WriteStringAt(25, 11, fmt.Sprintf("Max: %2.f km/h", maxWind))

	mntl.WriteStringAtWithAttributes(25, 13, fmt.Sprintf("Pression"), minigo.InversionFond)
	mntl.WriteStringAt(25, 14, fmt.Sprintf("Min: %d hPa", minPress))
	mntl.WriteStringAt(25, 15, fmt.Sprintf("Max: %d hPa", maxPress))

	mntl.WriteStringAtWithAttributes(25, 17, fmt.Sprintf("Ephéméride"), minigo.InversionFond)
	mntl.WriteStringAt(25, 18, fmt.Sprintf("Lev.: %s",
		time.Unix(forecast.City.Sunrise, 0).In(location).Format("15:04")))
	mntl.WriteStringAt(25, 19, fmt.Sprintf("Cou.: %s",
		time.Unix(forecast.City.Sunset, 0).In(location).Format("15:04")))

	return true
}
