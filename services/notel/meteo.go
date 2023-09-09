package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

var StationId = map[string]string{
	"07005": "ABBEVILLE",
	"07015": "LILLE",
	"07027": "CAEN",
	"07037": "ROUEN",
	"07072": "REIMS",
	"07110": "BREST",
	"07130": "RENNES",
	"07139": "ALENCON",
	"07149": "PARIS",
	"07168": "TROYES",
	"07181": "NANCY",
	"07190": "STRASBOURG",
	"07222": "NANTES",
	"07240": "TOURS",
	"07255": "BOURGES",
	"07280": "DIJON",
	"07299": "MULHOUSE",
	"07335": "POITIERS",
	"07434": "LIMOGES",
	"07460": "CLERMONT-FD",
	"07471": "LE PUY",
	"07481": "LYON",
	"07510": "BORDEAUX",
	"07535": "GOURDON",
	"07558": "MILLAU",
	"07577": "MONTELIMAR",
	"07607": "MONT-DE-MARSAN",
	"07621": "TARBES",
	"07630": "TOULOUSE",
	"07643": "MONTPELLIER",
	"07650": "MARSEILLE",
	"07690": "NICE",
	"07747": "PERPIGNAN",
	"07761": "AJACCIO",
	"07790": "BASTIA",
}

var PublicationsHours = []int{0, 3, 6, 9, 12, 15, 18, 21}

// ex: https://donneespubliques.meteofrance.fr/donnees_libres/Txt/Synop/synop.2023082315.csv
const FileFormat = "synop.%s.csv"
const URLFormat = "https://donneespubliques.meteofrance.fr/donnees_libres/Txt/Synop/%s"

const (
	stationIdCol   = 0
	pressureCol    = 2
	windSpeedCol   = 6
	temperatureCol = 7
	humidityCol    = 9
)

type WeatherReport struct {
	stationId   string
	stationName string
	temperature float64
	windSpeed   float64
	pressure    float64
	humidity    float64
}

func getLastWeatherData() ([]WeatherReport, error) {
	now := time.Now()

	lastPublicationHour := 3 * (now.UTC().Hour() / 3)
	lastPublicationDate := time.Date(now.Year(), now.Month(), now.Day(), lastPublicationHour, 0, 0, 0, time.UTC)

	fileName := fmt.Sprintf(FileFormat, lastPublicationDate.Format("2006010215"))
	filePath := fmt.Sprintf("/tmp/%s", fileName)
	fileURL := fmt.Sprintf(URLFormat, fileName)

	if _, err := os.Stat(filePath); err != nil {
		infoLog.Printf("weather file does not exist at: %s\n", filePath)
		infoLog.Printf("download file from: %s\n", fileURL)

		if err := downloadFile(fileURL, filePath); err != nil {
			errorLog.Printf("unable to download file: %s\n", err.Error())
			return nil, err
		}
	}

	weatherFile, err := os.Open(filePath)
	if err != nil {
		errorLog.Printf("unable to open file at: %s\n", filePath)
		return nil, err
	}
	defer weatherFile.Close()

	fileReader := csv.NewReader(weatherFile)
	fileReader.Comma = ';'

	weatheRecords, err := fileReader.ReadAll()
	if err != nil {
		errorLog.Printf("unable to read weather CSV record: %s\n", err.Error())
		return nil, err
	}

	weatherReports := []WeatherReport{}
	for _, record := range weatheRecords {
		stationName, ok := StationId[record[stationIdCol]]
		if ok {
			temp, err := strconv.ParseFloat(record[temperatureCol], 32)
			if err != nil {
				warnLog.Printf("unable to parse temperature for station %s: %s\n", stationName, err.Error())
				continue
			}

			pres, err := strconv.ParseFloat(record[pressureCol], 32)
			if err != nil {
				warnLog.Printf("unable to parse pressure for station %s: %s\n", stationName, err.Error())
				continue
			}

			hdty, err := strconv.ParseFloat(record[humidityCol], 32)
			if err != nil {
				warnLog.Printf("unable to parse humidity for station %s: %s\n", stationName, err.Error())
				continue
			}

			wind, err := strconv.ParseFloat(record[windSpeedCol], 32)
			if err != nil {
				warnLog.Printf("unable to parse wind-speed for station %s: %s\n", stationName, err.Error())
				continue
			}

			weatherReports = append(weatherReports, WeatherReport{
				stationId:   record[stationIdCol],
				stationName: stationName,
				temperature: temp,
				pressure:    pres,
				humidity:    hdty,
				windSpeed:   wind,
			})
		}
	}

	return weatherReports, nil
}

func downloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
