package main

import (
	"fmt"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/utils"
)

func NewObservationsPage(mntl *minigo.Minitel) *minigo.Page {
	meteoPage := minigo.NewPage("meteo", mntl, nil)

	// 5 lines, because
	// line=0   -- Reserved TELETEL
	// line=1   -- Date update
	// line=2   -- Blank
	// ...      -- Reports
	// line=22  -- Blank
	// line=23  -- Navigation Retour/Suite
	// line=24  -- Navigation Sommaire
	const reportsPerPage = (minigo.LignesSimple - 6) / 4
	maxPageId := 0
	pageId := 0

	var reports map[string][]WeatherReport

	meteoPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.CursorOff()

		var err error
		reports, err = getLastWeatherData()
		if err != nil {
			mntl.WriteStringAt(1, 1, "Connection à Météo-France echouée")
			mntl.WriteStringAt(2, 1, "Retour au sommaire dans 5 sec.")
			time.Sleep(5 * time.Second)
			return sommaireId
		}

		// pageId goes from 0 to maxPageId
		// e.g. for 3 items per page, and 9 items to display
		// 3/9 = 3, but maxPageId = 2
		// because pageId = [0, 1, 2] and len([0, 1, 2]) = 3
		maxPageId = len(reports)/reportsPerPage - 1

		printReportsFrom(mntl, reports, pageId, reportsPerPage, maxPageId)

		return minigo.NoOp
	})

	meteoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, sommaireId
	})

	meteoPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if pageId == maxPageId {
			return nil, minigo.NoOp
		}

		pageId += 1
		printReportsFrom(mntl, reports, pageId, reportsPerPage, maxPageId)
		return nil, minigo.NoOp
	})

	meteoPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if pageId == 0 {
			return nil, minigo.NoOp
		}

		pageId -= 1
		printReportsFrom(mntl, reports, pageId, reportsPerPage, maxPageId)
		return nil, minigo.NoOp
	})

	return meteoPage
}

func printReportsFrom(mntl *minigo.Minitel, reps map[string][]WeatherReport, pageId, reportsPerPage, maxPageId int) {
	mntl.CleanScreen()

	mntl.MoveAt(1, 1)
	mntl.WriteStringLeft(1, fmt.Sprintf("Mise à jour le: %s UTC", reps["07149"][0].date.Format("02/01/2006 15:04")))
	mntl.MoveAt(3, 1)

	for reportId := pageId * reportsPerPage; reportId < len(reps) && reportId < (pageId+1)*reportsPerPage; reportId += 1 {
		if rep, ok := reps[OrderedStationId[reportId]]; ok {
			printWeatherReport(mntl, rep)
		}
	}

	printHelpers(mntl, pageId, maxPageId)
}

func printWeatherReport(mntl *minigo.Minitel, reps []WeatherReport) {
	newestReport := reps[0]
	oldestReport := reps[0]

	for _, r := range reps {
		if r.date.Before(oldestReport.date) {
			oldestReport = r
		}

		if r.date.After(newestReport.date) {
			newestReport = r
		}
	}

	buf := minigo.EncodeAttributes(minigo.InversionFond)
	buf = append(buf, minigo.EncodeString(newestReport.stationName)...)
	buf = append(buf, minigo.EncodeAttributes(minigo.FondNormal)...)
	buf = append(buf, minigo.Return(1)...)

	// Temperature
	newTemp := newestReport.temperature - 275.
	difTemp := newestReport.temperature - oldestReport.temperature
	buf = append(buf, minigo.EncodeSprintf("%s%2.f°C (%+3.f°C) | ", utils.GetArrow(difTemp), newTemp, difTemp)...)

	// Pressure
	newPres := newestReport.pressure / 100.
	difPres := (newestReport.pressure - oldestReport.pressure) / 100.
	buf = append(buf, minigo.EncodeSprintf("%s%4.f hPa (%+5.f hPa)", utils.GetArrow(difPres), newPres, difPres)...)

	buf = append(buf, minigo.Return(1)...)

	// Wind
	arrow := utils.GetArrow(newestReport.windSpeed - oldestReport.windSpeed)
	buf = append(buf, minigo.EncodeSprintf(
		"%-10s %s%3.f km/h (%+3.f km/h)",
		windDirToString(newestReport.windDir),
		arrow,
		newestReport.windSpeed*3.6,
		(newestReport.windSpeed-oldestReport.windSpeed)*3.6)...)

	buf = append(buf, minigo.Return(2)...)

	mntl.Send(buf)
}

func printHelpers(mntl *minigo.Minitel, pageId, maxPageId int) {
	// pageId goes from 0 to maxPageId
	// PreviousPageNumber = (PageId + 1) - 1
	// NextPageNumber = (PageId + 1) + 1
	if pageId > 0 {
		mntl.WriteHelperLeft(23, fmt.Sprintf("Page %d/%d", pageId, maxPageId+1), "RETOUR")
	}
	if pageId < maxPageId {
		mntl.WriteHelperRight(23, fmt.Sprintf("Page %d/%d", pageId+2, maxPageId+1), "SUITE")
	}
	mntl.WriteHelperLeft(24, "Menu INFOMETEO", "SOMMAIRE")
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
