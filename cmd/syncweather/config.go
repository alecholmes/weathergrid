package main

type config struct {
	OpenWeatherMapAPIKey string           `json:"openweathermap_api_key"`
	SnapshotsBucketName  string           `json:"snapshots_bucket_name"`
	Groups               []*locationGroup `json:"groups"`
}

type locationGroup struct {
	Name      string      `json:"name"`
	Slug      string      `json:"slug"`
	Locations []*location `json:"locations"`
}

type location struct {
	Name string  `json:"name"`
	Slug string  `json:"slug"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}
