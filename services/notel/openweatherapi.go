package main

import "time"

const OWApiUrlFormat = "https://api.openweathermap.org/data/3.0/onecall?lat=%.5f&lon=%.5f&appid=%s&units=metric"

type OpenWeatherApiResponse struct {
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	Timezone       string  `json:"timezone"`
	TimezoneOffset int     `json:"timezone_offset"`
	Current        struct {
		Dt         time.Time `json:"dt"`
		Sunrise    int       `json:"sunrise"`
		Sunset     int       `json:"sunset"`
		Temp       float64   `json:"temp"`
		FeelsLike  float64   `json:"feels_like"`
		Pressure   int       `json:"pressure"`
		Humidity   int       `json:"humidity"`
		DewPoint   float64   `json:"dew_point"`
		Uvi        float64   `json:"uvi"`
		Clouds     int       `json:"clouds"`
		Visibility int       `json:"visibility"`
		WindSpeed  float64   `json:"wind_speed"`
		WindDeg    int       `json:"wind_deg"`
		WindGust   float64   `json:"wind_gust"`
		Weather    []struct {
			ID          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
	} `json:"current"`
	Minutely []struct {
		Dt            time.Time `json:"dt"`
		Precipitation int       `json:"precipitation"`
	} `json:"minutely"`
	Hourly []struct {
		Dt         time.Time `json:"dt"`
		Temp       float64   `json:"temp"`
		FeelsLike  float64   `json:"feels_like"`
		Pressure   int       `json:"pressure"`
		Humidity   int       `json:"humidity"`
		DewPoint   float64   `json:"dew_point"`
		Uvi        int       `json:"uvi"`
		Clouds     int       `json:"clouds"`
		Visibility int       `json:"visibility"`
		WindSpeed  float64   `json:"wind_speed"`
		WindDeg    int       `json:"wind_deg"`
		WindGust   float64   `json:"wind_gust"`
		Weather    []struct {
			ID          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Pop float64 `json:"pop"`
	} `json:"hourly"`
	Daily []struct {
		Dt        time.Time `json:"dt"`
		Sunrise   int       `json:"sunrise"`
		Sunset    int       `json:"sunset"`
		Moonrise  int       `json:"moonrise"`
		Moonset   int       `json:"moonset"`
		MoonPhase float64   `json:"moon_phase"`
		Summary   string    `json:"summary"`
		Temp      struct {
			Day   float64 `json:"day"`
			Min   float64 `json:"min"`
			Max   float64 `json:"max"`
			Night float64 `json:"night"`
			Eve   float64 `json:"eve"`
			Morn  float64 `json:"morn"`
		} `json:"temp"`
		FeelsLike struct {
			Day   float64 `json:"day"`
			Night float64 `json:"night"`
			Eve   float64 `json:"eve"`
			Morn  float64 `json:"morn"`
		} `json:"feels_like"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
		DewPoint  float64 `json:"dew_point"`
		WindSpeed float64 `json:"wind_speed"`
		WindDeg   int     `json:"wind_deg"`
		WindGust  float64 `json:"wind_gust"`
		Weather   []struct {
			ID          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Clouds int     `json:"clouds"`
		Pop    float64 `json:"pop"`
		Rain   float64 `json:"rain"`
		Uvi    float64 `json:"uvi"`
	} `json:"daily"`
	Alerts []struct {
		SenderName  string   `json:"sender_name"`
		Event       string   `json:"event"`
		Start       int      `json:"start"`
		End         int      `json:"end"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	} `json:"alerts"`
}
