package model

import "time"

type WeatherSnapshot struct {
	Version string           `json:"version"`
	Groups  []*LocationGroup `json:"groups"`
}

type LocationGroup struct {
	Name      string             `json:"name"`
	Slug      string             `json:"slug"`
	Locations []*LocationWeather `json:"locations"`
}

type LocationWeather struct {
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	Lat       float64    `json:"lat"`
	Lon       float64    `json:"lon"`
	Latest    *Weather   `json:"latest"`
	Forecasts []*Weather `json:"forecasts"`
}

type Weather struct {
	Timestamp            time.Time         `json:"timestamp"`
	Temp                 float64           `json:"temp"`
	TempFeelsLike        float64           `json:"temp_feels_like"`
	HumidityPercent      float64           `json:"humidity"`
	CloudinessPercent    float64           `json:"clouds"`
	WindSpeed            float64           `json:"wind_speed"`
	WindGustSpeed        float64           `json:"wind_gust"`
	WindDegree           float64           `json:"wind_degree"`
	PrecipitationPercent float64           `json:"precipitation_percent"`
	Details              []*WeatherDetails `json:"details"`
}

type WeatherDetails struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
}
