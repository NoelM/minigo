package main

import (
	"encoding/csv"
	"encoding/json"
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

type APIForecastReply struct {
	RequestState int    `json:"request_state"`
	RequestKey   string `json:"request_key"`
	Message      string `json:"message"`
	ModelRun     string `json:"model_run"`
	Source       string `json:"source"`
	Forecasts    map[string]Forecast
}

type Forecast struct {
	Temperature struct {
		TwoM       float64 `json:"2m"`
		Sol        float64 `json:"sol"`
		Five00HPa  float64 `json:"500hPa"`
		Eight50HPa float64 `json:"850hPa"`
	} `json:"temperature"`
	Pression struct {
		NiveauDeLaMer float64 `json:"niveau_de_la_mer"`
	} `json:"pression"`
	Pluie           float64 `json:"pluie"`
	PluieConvective float64 `json:"pluie_convective"`
	Humidite        struct {
		TwoM float64 `json:"2m"`
	} `json:"humidite"`
	VentMoyen struct {
		One0M float64 `json:"10m"`
	} `json:"vent_moyen"`
	VentRafales struct {
		One0M float64 `json:"10m"`
	} `json:"vent_rafales"`
	VentDirection struct {
		One0M float64 `json:"10m"`
	} `json:"vent_direction"`
	IsoZero     float64 `json:"iso_zero"`
	RisqueNeige string  `json:"risque_neige"`
	Cape        float64 `json:"cape"`
	Nebulosite  struct {
		Haute   float64 `json:"haute"`
		Moyenne float64 `json:"moyenne"`
		Basse   float64 `json:"basse"`
		Totale  float64 `json:"totale"`
	} `json:"nebulosite"`
}

func (f *APIForecastReply) UnmarshalJSON(d []byte) error {
	tmp := map[string]json.RawMessage{}
	err := json.Unmarshal(d, &tmp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(tmp["request_state"], &f.RequestState)
	if err != nil {
		return err
	}
	delete(tmp, "request_state")

	err = json.Unmarshal(tmp["request_key"], &f.RequestKey)
	if err != nil {
		return err
	}
	delete(tmp, "request_key")

	err = json.Unmarshal(tmp["message"], &f.Message)
	if err != nil {
		return err
	}
	delete(tmp, "message")

	err = json.Unmarshal(tmp["model_run"], &f.ModelRun)
	if err != nil {
		return err
	}
	delete(tmp, "model_run")

	err = json.Unmarshal(tmp["source"], &f.Source)
	if err != nil {
		return err
	}
	delete(tmp, "source")

	f.Forecasts = map[string]Forecast{}

	for k, v := range tmp {
		var item Forecast
		err := json.Unmarshal(v, &item)
		if err != nil {
			return err
		}

		f.Forecasts[k] = item
	}
	return nil
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
	const filePath = "/media/core/communes-departement-region.csv"

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
