package main

import (
	"time"

	"github.com/NoelM/minigo"
)

func NewObservationsPage(mntl *minigo.Minitel) *minigo.Page {
	meteoPage := minigo.NewPage("meteo", mntl, nil)

	const reportsPerPage = 24 / 4
	maxPageId := 0
	pageId := 0

	var reports map[string][]WeatherReport

	meteoPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		var err error
		reports, err = getLastWeatherData()
		if err != nil {
			mntl.WriteStringAt(1, 1, "CONNECTION A METEO-FRANCE ECHOUEE")
			mntl.WriteStringAt(1, 2, "RETOUR AU SOMMAIRE DANS 5 SEC.")
			time.Sleep(5 * time.Second)
			return sommaireId
		}

		maxPageId = len(reports) / reportsPerPage

		printReportsFrom(mntl, reports, pageId, reportsPerPage)

		return minigo.NoOp
	})

	meteoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	meteoPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		pageId += 1
		if pageId > maxPageId {
			pageId = maxPageId
		}
		printReportsFrom(mntl, reports, pageId, reportsPerPage)
		return nil, minigo.NoOp
	})

	meteoPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		pageId -= 1
		if pageId < 0 {
			pageId = 0
		}

		printReportsFrom(mntl, reports, pageId, reportsPerPage)
		return nil, minigo.NoOp
	})

	return meteoPage
}

func printReportsFrom(mntl *minigo.Minitel, reps map[string][]WeatherReport, pageId, reportsPerPage int) {
	mntl.CleanScreen()
	mntl.MoveCursorAt(1, 1)

	for reportId := pageId * reportsPerPage; reportId < len(reps) && reportId < (pageId+1)*reportsPerPage; reportId += 1 {
		if rep, ok := reps[OrderedStationId[reportId]]; ok {
			printWeatherReport(mntl, rep)
		}
	}
}

func printWeatherReport(mntl *minigo.Minitel, reps []WeatherReport) {
	newestReport := reps[0]
	oldestReport := reps[0]

	for _, r := range reps {
		if r.date.Before(oldestReport.date) {
			oldestReport = r
		} else if r.date.After(newestReport.date) {
			newestReport = r
		}
	}
	infoLog.Printf("min date %s\n", oldestReport.date.Format(time.RFC3339))
	infoLog.Printf("max date %s\n", newestReport.date.Format(time.RFC3339))

	buf := minigo.EncodeAttributes(minigo.InversionFond)
	buf = append(buf, minigo.EncodeMessage(newestReport.stationName)...)
	buf = append(buf, minigo.EncodeAttributes(minigo.FondNormal)...)
	buf = append(buf, minigo.GetMoveCursorRight(2)...)
	buf = append(buf, minigo.EncodeMessage(newestReport.date.Format("02/01/2006 A 15:04"))...)
	buf = append(buf, minigo.GetMoveCursorReturn(1)...)

	newTemp := newestReport.temperature - 275.
	difTemp := newestReport.temperature - oldestReport.temperature

	if difTemp > 0 {
		buf = append(buf, minigo.Ss2, minigo.FlecheHaut, minigo.Si)
	} else {
		buf = append(buf, minigo.Ss2, minigo.FlecheBas, minigo.Si)
	}
	buf = append(buf, minigo.EncodeSprintf(" %2.f", newTemp)...)
	buf = append(buf, minigo.EncodeSprintf(" (%2.f) ", difTemp)...)
	buf = append(buf, minigo.Ss2, minigo.Degre, minigo.Si)
	buf = append(buf, minigo.EncodeMessage("C")...)

	buf = append(buf, minigo.EncodeMessage(" - ")...)

	newPres := newestReport.pressure / 100.
	difPres := (newestReport.pressure - oldestReport.pressure) / 100.

	if difTemp > 0 {
		buf = append(buf, minigo.Ss2, minigo.FlecheHaut, minigo.Si)
	} else {
		buf = append(buf, minigo.Ss2, minigo.FlecheBas, minigo.Si)
	}
	buf = append(buf, minigo.EncodeSprintf(" %4.f", newPres)...)
	buf = append(buf, minigo.EncodeSprintf(" (%4.f) hPa", difPres)...)

	// wind
	buf = append(buf, minigo.GetMoveCursorReturn(1)...)
	buf = append(buf, minigo.EncodeMessage(windDirToString(newestReport.windDir))...)
	buf = append(buf, minigo.GetMoveCursorRight(1)...)

	if newestReport.windSpeed-oldestReport.windSpeed > 0 {
		buf = append(buf, minigo.Ss2, minigo.FlecheHaut, minigo.Si)
	} else {
		buf = append(buf, minigo.Ss2, minigo.FlecheBas, minigo.Si)
	}
	buf = append(buf, minigo.EncodeSprintf(" %3.f (%3.f) km/h", newestReport.windSpeed*3.6, (newestReport.windSpeed-oldestReport.windSpeed)*3.6)...)

	buf = append(buf, minigo.GetMoveCursorReturn(2)...)

	mntl.Send(buf)
}

func windDirToString(deg float64) string {
	if deg > 337.5 || deg <= 22.5 {
		return "NORD"
	} else if deg > 22.5 && deg <= 67.5 {
		return "NORD-EST"
	} else if deg > 67.5 && deg <= 112.5 {
		return "EST"
	} else if deg > 112.5 && deg <= 157.5 {
		return "SUD-EST"
	} else if deg > 157.5 && deg <= 202.5 {
		return "SUD"
	} else if deg > 202.5 && deg <= 247.5 {
		return "SUD-OUEST"
	} else if deg > 247.5 && deg <= 292.5 {
		return "OUEST"
	} else if deg > 292.5 && deg <= 337.5 {
		return "NORD-OUEST"
	} else {
		return ""
	}
}
