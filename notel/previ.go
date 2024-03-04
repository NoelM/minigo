package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"sync"

	"github.com/NoelM/minigo/notel/logs"
)

type Commune struct {
	CodeCommune     string  `json:"code_commune"`
	NomCommune      string  `json:"nom_commune"`
	CodePostal      string  `json:"code_postal"`
	CodeDepartement string  `json:"code_departement"`
	NomDepartement  string  `json:"nom_departement"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
}

type CommuneDatabase struct {
	mutex               sync.RWMutex
	CodePostalToCommune map[string][]Commune
}

func NewCommuneDatabase() *CommuneDatabase {
	return &CommuneDatabase{
		CodePostalToCommune: make(map[string][]Commune),
	}
}

func (c *CommuneDatabase) GetCommunesFromCodePostal(codePostal string) []Commune {
	if codePostal[0] == '0' {
		codePostal = codePostal[1:]
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	communes, ok := c.CodePostalToCommune[codePostal]
	if !ok {
		return nil
	}

	return communes
}

func (c *CommuneDatabase) LoadCommuneDatabase(filePath string) error {
	const codeCommuneColId = 0
	const nomCommuneColId = 10
	const codePostalColId = 2
	const latitudeColId = 5
	const longitudeColId = 6
	const codeDepartementColId = 11
	const nomDepartementColId = 12

	communesFile, err := os.Open(filePath)
	if err != nil {
		logs.ErrorLog("unable to open file at: %s\n", filePath)
		return err
	}
	defer communesFile.Close()

	fileReader := csv.NewReader(communesFile)
	fileReader.Comma = ','

	communeRecords, err := fileReader.ReadAll()
	if err != nil {
		logs.ErrorLog("unable to read commune CSV records: %s\n", err.Error())
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, record := range communeRecords {
		communeName := record[nomCommuneColId]
		codePostal := record[codePostalColId]
		codeDepartement := record[codeDepartementColId]

		parsedDepCode, err := strconv.ParseInt(codeDepartement, 10, 32)
		if err != nil && codeDepartement != "2A" && codeDepartement != "2B" {
			logs.WarnLog("unable to parse code departement for commune %s: %s\n", communeName, err.Error())
			continue
		} else if parsedDepCode > 95 {
			//logs.InfoLog("ignore commune %s because out of metropole\n", communeName)
			continue
		}

		lat, err := strconv.ParseFloat(record[latitudeColId], 32)
		if err != nil {
			logs.WarnLog("unable to parse latitude for commune %s: %s\n", communeName, err.Error())
			continue
		}

		lon, err := strconv.ParseFloat(record[longitudeColId], 32)
		if err != nil {
			logs.WarnLog("unable to parse longitude for commune %s: %s\n", communeName, err.Error())
			continue
		}

		_, ok := c.CodePostalToCommune[codePostal]
		if !ok {
			c.CodePostalToCommune[codePostal] = []Commune{
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
			c.CodePostalToCommune[codePostal] = append(c.CodePostalToCommune[codePostal], Commune{
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
