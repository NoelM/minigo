package main

import "time"

const OWApiUrlFormat = "https://api.openweathermap.org/data/2.5/forecast?lat=%.5f&lon=%.5f&appid=%s&units=metric"

type OpenWeatherApiResponse struct {
	Cod     string `json:"cod"`
	Message int    `json:"message"`
	Cnt     int    `json:"cnt"`
	List    []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			TempMin   float64 `json:"temp_min"`
			TempMax   float64 `json:"temp_max"`
			Pressure  int     `json:"pressure"`
			SeaLevel  int     `json:"sea_level"`
			GrndLevel int     `json:"grnd_level"`
			Humidity  int     `json:"humidity"`
			TempKf    float64 `json:"temp_kf"`
		} `json:"main"`
		Weather []struct {
			ID          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Clouds struct {
			All int `json:"all"`
		} `json:"clouds"`
		Wind struct {
			Speed float64 `json:"speed"`
			Deg   int     `json:"deg"`
			Gust  float64 `json:"gust"`
		} `json:"wind"`
		Visibility int     `json:"visibility"`
		Pop        float64 `json:"pop"`
		Rain       struct {
			ThreeH float64 `json:"3h"`
		} `json:"rain,omitempty"`
		Sys struct {
			Pod string `json:"pod"`
		} `json:"sys"`
		DtTxt string `json:"dt_txt"`
	} `json:"list"`
	City struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Coord struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"coord"`
		Country    string `json:"country"`
		Population int    `json:"population"`
		Timezone   int    `json:"timezone"`
		Sunrise    int64  `json:"sunrise"`
		Sunset     int64  `json:"sunset"`
	} `json:"city"`
}

func weatherConditionCodeToString(code int, forecastDate time.Time) string {
	switch code / 100 {
	case 2:
		return "Orage"
	case 3:
		return "Bruine"
	case 5:
		return "Pluie"
	case 6:
		return "Neige"
	case 7:
		return "Fumées"
	case 8:
		if code == 800 || code == 801 {
			if forecastDate.Hour() > 7 && forecastDate.Hour() < 19 {
				return "Soleil"
			} else {
				return "Clair"
			}
		} else if code == 802 {
			return "Qlq Nuages"
		} else if code == 803 {
			return "Nuageux"
		} else if code == 804 {
			return "Couvert"
		}
	}
	return ""
}
