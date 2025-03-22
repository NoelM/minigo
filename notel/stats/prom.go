package stats

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/NoelM/minigo/notel/logs"
)

type PromApiResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
			} `json:"metric"`
			Value []any `json:"value"`
		} `json:"result"`
	} `json:"data"`
	Warnings []string `json:"warnings"`
}

const ConnectWeekly = `sum(increase(notel_connection_number{job="notel"}[1w]))`
const DurationWeekly = `sum(increase(notel_connection_duration{job="notel"}[1w]))`
const MessagesWeekly = `sum(increase(notel_messages_number{job="notel"}[1w]))`
const CPUTemp = `node_thermal_zone_temp{zone="0"}`
const CPULoad = `100 - (avg by (instance) (rate(node_cpu_seconds_total{job="node",mode="idle"}[5m])) * 100)`

func Request(query string) (r PromApiResponse, err error) {
	promURL := os.Getenv("PROM_URL")
	promUser := os.Getenv("PROM_USER")
	promPass := os.Getenv("PROM_PASS")

	client := &http.Client{}

	data := url.Values{}
	data.Set("query", query)

	fullUrl, _ := url.Parse(promURL)
	fullUrl.RawQuery = data.Encode()

	// Set up HTTPS request with basic authorization.
	req, err := http.NewRequest(http.MethodGet, fullUrl.String(), nil)
	if err != nil {
		logs.ErrorLog("unable to build prom request: %s\n", err.Error())
		return
	}
	req.SetBasicAuth(promUser, promPass)

	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorLog("unable to request prom: %s\n", err.Error())
		return
	}
	defer resp.Body.Close()

	promData, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorLog("unable to read reply: %s\n", err.Error())
		return
	}
	fmt.Println(string(promData))

	if err = json.Unmarshal(promData, &r); err != nil {
		logs.ErrorLog("unable to unmarshal data: %s\n%s", err.Error(), promData)
		return
	}

	return
}
