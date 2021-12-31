package openweathermap

type OnecallResponse struct {
	Lat            float64    `json:"lat"`
	Lon            float64    `json:"lon"`
	Timezone       string     `json:"timezone"`
	TimezoneOffset int        `json:"timezone_offset"`
	Current        *Weather   `json:"current"`
	Hourly         []*Weather `json:"hourly"`
}

type Weather struct {
	Timestamp            int64             `json:"dt"`
	Temp                 float64           `json:"temp"`
	TempFeelsLike        float64           `json:"feels_like"`
	HumidityPercent      float64           `json:"humidity"`
	CloudinessPercent    float64           `json:"clouds"`
	WindSpeed            float64           `json:"wind_speed"`
	WindGustSpeed        float64           `json:"wind_gust"`
	WindDegree           float64           `json:"wind_deg"`
	PrecipitationPercent float64           `json:"pop"`
	Details              []*WeatherDetails `json:"weather"`
}

type WeatherDetails struct {
	Main        string `json:"main"`
	Description string `json:"description"`
}
