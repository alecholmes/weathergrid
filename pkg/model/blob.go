package model

type WeatherBlob struct {
	Locations []*LocationWeather `json:"locations"`
}

type LocationWeather struct {
	Name string `json:"name"`
}
