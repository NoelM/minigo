package meteo

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/NoelM/minigo/notel/logs"
	"github.com/NoelM/minigo/notel/utils"
)

func NewPrevisionPage(mntl *minigo.Minitel, communeMap map[string]string) *minigo.Page {
	previPage := minigo.NewPage("previsions", mntl, communeMap)

	var forecast OpenWeatherApiResponse
	var commune databases.Commune

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
		mntl.Reset()
		mntl.CursorOff()

		communeJSON, ok := initData["commune"]
		if !ok {
			logs.ErrorLog("no commune data\n")
			return minigo.SommaireOp
		}

		if err := json.Unmarshal([]byte(communeJSON), &commune); err != nil {
			logs.ErrorLog("unable to parse the commune JSON: %s\n", err.Error())
			return minigo.SommaireOp
		}

		OWApiKey := os.Getenv("OWAPIKEY")
		body, err := getRequestBody(fmt.Sprintf(OWApiUrlFormat, commune.Latitude, commune.Longitude, OWApiKey))
		if err != nil {
			logs.ErrorLog("unable to get forecasts: %s\n", err.Error())
			return minigo.SommaireOp
		}
		defer body.Close()

		data, err := io.ReadAll(body)
		if err != nil {
			logs.ErrorLog("unable to get API response: %s\n", err.Error())
			return minigo.SommaireOp
		}

		if err := json.Unmarshal(data, &forecast); err != nil {
			logs.ErrorLog("unable to parse JSON: %s\n", err.Error())
			return minigo.SommaireOp
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
		return nil, minigo.SommaireOp
	})

	return previPage
}

func printPreviHelpers(mntl *minigo.Minitel, forecastDate, firstForecastDate, lastForecastDate time.Time) {
	mntl.MoveAt(23, 0)

	if forecastDate.After(firstForecastDate) {
		mntl.Helper(forecastDate.Add(-24*time.Hour).Format("02/01"), "RETOUR", minigo.FondBleu, minigo.CaractereBlanc)
	}
	if forecastDate.Before(lastForecastDate) {
		mntl.HelperRight(forecastDate.Add(24*time.Hour).Format("02/01"), "SUITE", minigo.FondBleu, minigo.CaractereBlanc)
	}
	mntl.Return(1)
	mntl.HelperRight("Menu InfoMétéo", "SOMMAIRE", minigo.FondCyan, minigo.CaractereNoir)
}

func printForecast(mntl *minigo.Minitel, forecast OpenWeatherApiResponse, forecastDate time.Time, c databases.Commune) {
	mntl.CleanScreen()
	location, _ := time.LoadLocation("Europe/Paris")

	// City name
	mntl.MoveAt(2, 1)
	mntl.Attributes(minigo.DoubleHauteur)
	if len(c.NomCommune) >= minigo.ColonnesSimple {
		mntl.Print(c.NomCommune[:minigo.ColonnesSimple-1])
	} else {
		mntl.Print(c.NomCommune)
	}
	mntl.Attributes(minigo.GrandeurNormale)

	// Date of the forecast
	mntl.ReturnCol(1, 1)
	mntl.Print(fmt.Sprintf("%s %d %s",
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
	mntl.ReturnCol(2, 1)
	for fDate := forecastDate; fDate.Before(forecastDate.Add(24 * time.Hour)); fDate = fDate.Add(3 * time.Hour) {
		for _, fct := range forecast.List {
			hourlyDate := time.Unix(fct.Dt, 0)
			if hourlyDate.After(fDate.Add(-time.Hour)) && hourlyDate.Before(fDate.Add(time.Hour)) {
				previsionString := fmt.Sprintf("%s> %2.f°C %s",
					fDate.In(location).Format("15:04"),
					fct.Main.Temp,
					weatherConditionCodeToString(fct.Weather[0].ID, fDate.In(location)))

				mntl.Print(previsionString)

				mntl.ReturnCol(2, 1)
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

	// TEMPERATURES
	mntl.MoveAt(5, 25)
	mntl.PrintAttributes("Températures", minigo.InversionFond)
	mntl.MoveOf(1, -12)
	mntl.Print(fmt.Sprintf("Min: %2.f°C", minTemp))
	mntl.MoveOf(1, -9)
	mntl.Print(fmt.Sprintf("Max: %2.f°C", maxTemp))

	mntl.MoveOf(2, -9)
	mntl.PrintAttributes("Vent", minigo.InversionFond)
	mntl.MoveOf(1, -4)
	mntl.Print(fmt.Sprintf("Min: %2.f km/h", 3.6*minWind))
	mntl.MoveOf(1, -12)
	mntl.Print(fmt.Sprintf("Max: %2.f km/h", 3.6*maxWind))

	mntl.MoveOf(2, -12)
	mntl.PrintAttributes("Pression", minigo.InversionFond)
	mntl.MoveOf(1, -8)
	mntl.Print(fmt.Sprintf("Min: %4d hPa", minPress))
	mntl.MoveOf(1, -13)
	mntl.Print(fmt.Sprintf("Max: %4d hPa", maxPress))

	mntl.MoveOf(2, -13)
	mntl.PrintAttributes("Ephéméride", minigo.InversionFond)
	mntl.MoveOf(1, -10)
	mntl.Print(fmt.Sprintf("Lev.: %s",
		time.Unix(forecast.City.Sunrise, 0).In(location).Format("15:04")))
	mntl.MoveOf(1, -11)
	mntl.Print(fmt.Sprintf("Cou.: %s",
		time.Unix(forecast.City.Sunset, 0).In(location).Format("15:04")))
}
