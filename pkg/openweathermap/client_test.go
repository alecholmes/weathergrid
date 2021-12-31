package openweathermap

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeh(t *testing.T) {
	// -28800

	ts := time.Unix(1640581527, 0)

	ftz := time.FixedZone("fz", -28800)
	ntz, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)

	fmt.Println(ts.Format(time.RFC3339))
	fmt.Println(ts.In(ftz).Format(time.RFC3339))
	fmt.Println(ts.In(ntz).Format(time.RFC3339))
}

func TestOnecall(t *testing.T) {
	testAPIKey := "test-api-key"

	client := NewRealClient(testAPIKey)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, onecallPath, r.URL.Path)
		assert.Equal(t, testAPIKey, r.URL.Query().Get("appid"))

		_, err := w.Write([]byte(testOnecallResp))
		require.NoError(t, err)
	}))
	defer server.Close()

	var err error
	client.BaseURL, err = url.Parse(server.URL)
	require.NoError(t, err)

	resp, err := client.OneCall(context.Background(), 1.11, 2.22, []ExcludeType{ExcludeMinutely, ExcludeAlerts}, UnitsImperial)
	require.NoError(t, err)

	assert.Equal(t, &OnecallResponse{
		Lat:            37.7693,
		Lon:            -122.4332,
		Timezone:       "America/Los_Angeles",
		TimezoneOffset: -28800,
		Current: &Weather{
			Timestamp:         1640564736,
			Temp:              45.97,
			TempFeelsLike:     44.56,
			HumidityPercent:   86,
			CloudinessPercent: 90,
			WindSpeed:         1.01,
			WindGustSpeed:     4,
			WindDegree:        183,
			Details: []*WeatherDetails{
				{
					Main:        "Rain",
					Description: "light rain",
				},
			},
		},
		Hourly: []*Weather{
			{
				Timestamp:            1640563200,
				Temp:                 45.97,
				TempFeelsLike:        38.91,
				HumidityPercent:      86,
				CloudinessPercent:    90,
				WindSpeed:            17.25,
				WindGustSpeed:        21.59,
				WindDegree:           230,
				PrecipitationPercent: 0.47,
				Details: []*WeatherDetails{
					{
						Main:        "Clouds",
						Description: "overcast clouds",
					},
				},
			},
		},
	}, resp)
}

var testOnecallResp = `{
  "lat": 37.7693,
  "lon": -122.4332,
  "timezone": "America/Los_Angeles",
  "timezone_offset": -28800,
  "current": {
    "dt": 1640564736,
    "sunrise": 1640532223,
    "sunset": 1640566626,
    "temp": 45.97,
    "feels_like": 44.56,
    "pressure": 1017,
    "humidity": 86,
    "dew_point": 42.03,
    "uvi": 0.07,
    "clouds": 90,
    "visibility": 10000,
    "wind_speed": 1.01,
    "wind_deg": 183,
    "wind_gust": 4,
    "weather": [
      {
        "id": 500,
        "main": "Rain",
        "description": "light rain",
        "icon": "10d"
      }
    ],
    "rain": {
      "1h": 0.75
    }
  },
  "hourly": [
    {
      "dt": 1640563200,
      "temp": 45.97,
      "feels_like": 38.91,
      "pressure": 1017,
      "humidity": 86,
      "dew_point": 42.03,
      "uvi": 0.07,
      "clouds": 90,
      "visibility": 10000,
      "wind_speed": 17.25,
      "wind_deg": 230,
      "wind_gust": 21.59,
      "weather": [
        {
          "id": 804,
          "main": "Clouds",
          "description": "overcast clouds",
          "icon": "04d"
        }
      ],
      "pop": 0.47
    }
  ]
}`
