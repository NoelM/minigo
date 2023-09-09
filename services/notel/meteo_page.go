package main

import (
	"time"

	"github.com/NoelM/minigo"
)

const reportSizeInLines = 2
const maxReportsPerPage = 25 / reportSizeInLines

func NewMeteoPage(mntl *minigo.Minitel) *minigo.Page {
	meteoPage := minigo.NewPage("meteo", mntl, nil)

	currentReportId := 0
	var reports []WeatherReport

	meteoPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		var err error
		reports, err = getLastWeatherData()
		if err != nil {
			mntl.WriteStringXY(1, 1, "CONNECTION A METEO-FRANCE ECHOUEE")
			mntl.WriteStringXY(1, 1, "RETOUR AU SOMMAIRE DANS 5 SEC.")
			time.Sleep(5 * time.Second)
			return sommaireId
		}

		currentReportId = printReportsFrom(mntl, reports, currentReportId)

		return minigo.NoOp
	})

	meteoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	meteoPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		currentReportId = printReportsFrom(mntl, reports, currentReportId)
		return nil, minigo.NoOp
	})

	meteoPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		currentReportId -= 2 * maxReportsPerPage
		if currentReportId < 0 {
			currentReportId = 0
		}

		currentReportId = printReportsFrom(mntl, reports, currentReportId)
		return nil, minigo.NoOp
	})

	return meteoPage
}

func printReportsFrom(mntl *minigo.Minitel, reps []WeatherReport, from int) int {
	mntl.CleanScreen()
	mntl.MoveCursorXY(1, 1)

	id := from
	numberOfReports := 0
	for ; id < len(reps) && numberOfReports < maxReportsPerPage; id += 1 {
		printWeatherReport(mntl, reps[id])
		numberOfReports += 1
	}

	return id
}

func printWeatherReport(mntl *minigo.Minitel, rep WeatherReport) {
	buf := minigo.EncodeAttributes(minigo.InversionFond)
	buf = append(buf, minigo.EncodeMessage(rep.stationName)...)
	buf = append(buf, minigo.EncodeAttributes(minigo.FondNormal)...)
	buf = append(buf, minigo.GetMoveCursorReturn(1)...)

	// len 32 chars
	buf = append(buf, minigo.EncodeSprintf("%2.f C %3.f %% - %4.f hPa - %s %3.f km/h", rep.temperature-275., rep.humidity, rep.pressure/100., windDirToString(rep.windDir), rep.windSpeed*3.6)...)
	buf = append(buf, minigo.GetMoveCursorReturn(1)...)

	mntl.Send(buf)
}

func windDirToString(deg float64) string {
	if deg > 337.5 && deg <= 22.5 {
		return "N"
	} else if deg > 22.5 && deg <= 67.5 {
		return "NE"
	} else if deg > 67.5 && deg <= 112.5 {
		return "E"
	} else if deg > 112.5 && deg <= 157.5 {
		return "SE"
	} else if deg > 157.5 && deg <= 202.5 {
		return "S"
	} else if deg > 202.5 && deg <= 247.5 {
		return "SO"
	} else if deg > 247.5 && deg <= 292.5 {
		return "O"
	} else if deg > 292.5 && deg <= 337.5 {
		return "NO"
	} else {
		return ""
	}
}
