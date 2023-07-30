package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type Response struct {
	Departures []Departure `json:"departures"`
	Ctx        Context     `json:"context"`
	Mtx        sync.RWMutex
}

type Departure struct {
	Informations DisplayInformation `json:"display_informations"`
	Schedule     StopDateTime       `json:"stop_date_time"`
}

type DisplayInformation struct {
	Code           string `json:"code"`
	CommercialMode string `json:"commercial_mode"`
	Direction      string `json:"direction"`
	Headsign       string `json:"headsign"`
}

type StopDateTime struct {
	BaseDepartureDateTime string `json:"base_departure_date_time"`
	DepartureDateTime     string `json:"departure_date_time"`
}

type Context struct {
	CurrentDatetime string `json:"current_datetime"`
	Timezone        string `json:"timezone"`
}

func GetDepartures(apiKey string, resp *Response) error {
	requestURL := fmt.Sprintf("https://api.navitia.io/v1/coverage/sncf/stop_areas/stop_area%%3ASNCF%%3A87686006/physical_modes/physical_mode%%3ALongDistanceTrain/departures?from_datetime=%s&", time.Now().Format("20060102T150405"))

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("client: could not create request: %s", err)
	}

	req.Header.Add("Authorization", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("client: error making http request: %s", err)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("client: could not read response body: %s", err)
	}

	resp.Mtx.Lock()
	defer resp.Mtx.Unlock()

	err = json.Unmarshal(resBody, resp)
	return err
}
