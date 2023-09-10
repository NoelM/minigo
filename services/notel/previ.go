package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"sync"
)

const LatColId = 5
const LonColId = 6

type Commune struct {
	CodeCommune     string  `json:"codeCommune"`
	NomCommune      string  `json:"nomCommune"`
	CodePostal      string  `json:"codePostal"`
	CodeDepartement string  `json:"codeDepartement"`
	NomDepartement  string  `json:"nomDepartement"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
}

func getCommunesFromCodePostal(codePostal string) []Commune {
	if codePostal[0] == '0' {
		codePostal = codePostal[1:]
	}

	CommuneDatabase.mutex.RLock()
	defer CommuneDatabase.mutex.RUnlock()

	communes, ok := CommuneDatabase.CodePostalToCommune[codePostal]
	if !ok {
		return nil
	}

	return communes
}

type CommuneDb struct {
	mutex               sync.RWMutex
	CodePostalToCommune map[string][]Commune
}

func loadCommuneDatabase() error {
	const filePath = "/media/code/commune-departement-region.csv"

	const codeCommuneColId = 0
	const nomCommuneColId = 10
	const codePostalColId = 2
	const latitudeColId = 5
	const longitudeColId = 6
	const codeDepartementColId = 11
	const nomDepartementColId = 12

	communesFile, err := os.Open(filePath)
	if err != nil {
		errorLog.Printf("unable to open file at: %s\n", filePath)
		return err
	}
	defer communesFile.Close()

	fileReader := csv.NewReader(communesFile)
	fileReader.Comma = ','

	communeRecords, err := fileReader.ReadAll()
	if err != nil {
		errorLog.Printf("unable to read commune CSV records: %s\n", err.Error())
		return err
	}

	CommuneDatabase.CodePostalToCommune = make(map[string][]Commune)
	CommuneDatabase.mutex.Lock()
	defer CommuneDatabase.mutex.Unlock()

	for _, record := range communeRecords {
		communeName := record[nomCommuneColId]
		codePostal := record[codePostalColId]
		codeDepartement := record[codeDepartementColId]

		parsedDepCode, err := strconv.ParseInt(codeDepartement, 10, 32)
		if err != nil && codeDepartement != "2A" && codeDepartement != "2B" {
			warnLog.Printf("unable to parse code departement for commune %s: %s\n", communeName, err.Error())
			continue
		} else if parsedDepCode > 95 {
			infoLog.Printf("ignore commune %s because out of metropole\n", communeName)
			continue
		}

		lat, err := strconv.ParseFloat(record[latitudeColId], 32)
		if err != nil {
			warnLog.Printf("unable to parse latitude for commune %s: %s\n", communeName, err.Error())
			continue
		}

		lon, err := strconv.ParseFloat(record[longitudeColId], 32)
		if err != nil {
			warnLog.Printf("unable to parse longitude for commune %s: %s\n", communeName, err.Error())
			continue
		}

		_, ok := CommuneDatabase.CodePostalToCommune[codePostal]
		if !ok {
			CommuneDatabase.CodePostalToCommune[codePostal] = []Commune{
				{
					CodeCommune:     record[codeCommuneColId],
					NomCommune:      communeName,
					CodePostal:      codePostal,
					CodeDepartement: codeDepartement,
					NomDepartement:  record[nomDepartementColId],
					Latitude:        lat,
					Longitude:       lon,
				},
			}
		} else {
			CommuneDatabase.CodePostalToCommune[codePostal] = append(CommuneDatabase.CodePostalToCommune[codePostal], Commune{
				CodeCommune:     record[codeCommuneColId],
				NomCommune:      communeName,
				CodePostal:      codePostal,
				CodeDepartement: codeDepartement,
				NomDepartement:  record[nomDepartementColId],
				Latitude:        lat,
				Longitude:       lon,
			})
		}
	}

	return nil
}
