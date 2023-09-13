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

var StationIdToName = map[string]string{
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

var OrderedStationId = []string{
	"07005", //"ABBEVILLE",
	"07761", //"AJACCIO",
	"07139", //"ALENCON",
	"07790", //"BASTIA",
	"07510", //"BORDEAUX",
	"07255", //"BOURGES",
	"07110", //"BREST",
	"07027", //"CAEN",
	"07460", //"CLERMONT-FD",
	"07280", //"DIJON",
	"07535", //"GOURDON",
	"07015", //"LILLE",
	"07434", //"LIMOGES",
	"07481", //"LYON",
	"07650", //"MARSEILLE",
	"07558", //"MILLAU",
	"07607", //"MONT-DE-MARSAN",
	"07577", //"MONTELIMAR",
	"07643", //"MONTPELLIER",
	"07299", //"MULHOUSE",
	"07181", //"NANCY",
	"07222", //"NANTES",
	"07690", //"NICE",
	"07149", //"PARIS",
	"07747", //"PERPIGNAN",
	"07335", //"POITIERS",
	"07471", //"LE PUY",
	"07072", //"REIMS",
	"07130", //"RENNES",
	"07037", //"ROUEN",
	"07190", //"STRASBOURG",
	"07621", //"TARBES",
	"07630", //"TOULOUSE",
	"07240", //"TOURS",
	"07168", //"TROYES",
}

var PublicationsHours = []int{0, 3, 6, 9, 12, 15, 18, 21}

// ex: https://donneespubliques.meteofrance.fr/donnees_libres/Txt/Synop/synop.2023082315.csv
const FileFormat = "synop.%s.csv"
const URLFormat = "https://donneespubliques.meteofrance.fr/donnees_libres/Txt/Synop/%s"

const (
	stationIdCol   = 0
	dateIdCol      = 1
	pressureCol    = 2
	windDirCol     = 5
	windSpeedCol   = 6
	temperatureCol = 7
	humidityCol    = 9
)

type WeatherReport struct {
	stationId   string
	stationName string
	date        time.Time
	temperature float64
	windDir     float64
	windSpeed   float64
	pressure    float64
	humidity    float64
}

func getLastWeatherData() (map[string][]WeatherReport, error) {
	// prevent the case of request the file exactly when produced
	now := time.Now().Add(-15 * time.Minute)

	lastPublicationHour := 3 * (now.UTC().Hour() / 3)
	lastPublicationDate := time.Date(now.Year(), now.Month(), now.Day(), lastPublicationHour, 0, 0, 0, time.UTC)

	globalReports := make(map[string][]WeatherReport)

	// 8 files by 24h
	for i := 0; i < 8; i += 1 {
		var filePath string
		var err error

		fileName := fmt.Sprintf(FileFormat, lastPublicationDate.Add(-3*time.Hour*time.Duration(i)).Format("2006010215"))
		if filePath, err = downloadFileIfDoesNotExist(fileName); err != nil {
			return nil, err
		}

		if err = openAndParseFile(filePath, globalReports); err != nil {
			errorLog.Printf("ignored file: %s: %s\n", filePath, err.Error())
			continue
		}
	}

	return globalReports, nil
}

func downloadFileIfDoesNotExist(fileName string) (string, error) {
	filePath := fmt.Sprintf("/media/core/%s", fileName)
	fileURL := fmt.Sprintf(URLFormat, fileName)

	if _, err := os.Stat(filePath); err != nil {
		infoLog.Printf("weather file does not exist at: %s\n", filePath)
		infoLog.Printf("download file from: %s\n", fileURL)

		if err := downloadFile(fileURL, filePath); err != nil {
			errorLog.Printf("unable to download file: %s\n", err.Error())
			return "", err
		}
	}
	return filePath, nil
}

func openAndParseFile(filePath string, globalReports map[string][]WeatherReport) error {
	weatherFile, err := os.Open(filePath)
	if err != nil {
		errorLog.Printf("unable to open file at: %s\n", filePath)
		return err
	}
	defer weatherFile.Close()

	fileReader := csv.NewReader(weatherFile)
	fileReader.Comma = ';'

	weatheRecords, err := fileReader.ReadAll()
	if err != nil {
		errorLog.Printf("unable to read weather CSV record: %s\n", err.Error())
		return err
	}

	for _, record := range weatheRecords {
		stationId := record[stationIdCol]
		stationName, ok := StationIdToName[stationId]
		if ok {
			dte, err := time.Parse("20060102150405", record[dateIdCol])
			if err != nil {
				warnLog.Printf("unable to parse date for station %s: %s\n", stationName, err.Error())
				continue
			}

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

			windDir, err := strconv.ParseFloat(record[windDirCol], 32)
			if err != nil {
				warnLog.Printf("unable to parse wind-dir for station %s: %s\n", stationName, err.Error())
				continue
			}

			windSpeed, err := strconv.ParseFloat(record[windSpeedCol], 32)
			if err != nil {
				warnLog.Printf("unable to parse wind-speed for station %s: %s\n", stationName, err.Error())
				continue
			}

			rep := WeatherReport{
				stationId:   record[stationIdCol],
				stationName: stationName,
				date:        dte,
				temperature: temp,
				pressure:    pres,
				humidity:    hdty,
				windSpeed:   windSpeed,
				windDir:     windDir,
			}

			if stationReports, ok := globalReports[stationId]; ok {
				stationReports = append(stationReports, rep)
			} else {
				globalReports[stationId] = []WeatherReport{rep}
			}
		}
	}

	return nil
}

func getRequestBody(url string) (io.ReadCloser, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func downloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	body, err := getRequestBody(url)
	if err != nil {
		return err
	}
	defer body.Close()

	// Write the body to file
	_, err = io.Copy(out, body)
	if err != nil {
		return err
	}

	return nil
}
