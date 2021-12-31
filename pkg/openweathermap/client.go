package openweathermap

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type ExcludeType string
type Units string

const (
	ExcludeCurrent  ExcludeType = "current"
	ExcludeMinutely             = "minutely"
	ExcludeHourly               = "hourly"
	ExcludeDaily                = "daily"
	ExcludeAlerts               = "alerts"

	UnitsStandard Units = "standard"
	UnitsMetric         = "metric"
	UnitsImperial       = "imperial"

	onecallPath = "/data/2.5/onecall"
)

var (
	baseURL *url.URL
)

func init() {
	var err error
	baseURL, err = url.Parse("https://api.openweathermap.org")
	if err != nil {
		panic(err)
	}
}

type Client interface {
	OneCall(
		ctx context.Context,
		lat, long float64,
		exclude []ExcludeType,
		units Units,
	) (*OnecallResponse, error)
}

func NewRealClient(apiKey string) *RealClient {
	return &RealClient{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: http.DefaultClient,
	}
}

type RealClient struct {
	APIKey     string
	BaseURL    *url.URL
	HTTPClient *http.Client
}

var _ Client = (*RealClient)(nil)

// OneCall gets weather data for a specific location. See https://openweathermap.org/api/one-call-api
func (r *RealClient) OneCall(ctx context.Context, lat, lon float64, exclude []ExcludeType, units Units) (*OnecallResponse, error) {
	header := make(http.Header, 2)
	header.Set("content-type", "application/json")
	header.Set("accept", "application/json")

	reqURL := *r.BaseURL
	reqURL.Path += onecallPath

	query := make(url.Values)
	query.Set("lat", fmt.Sprintf("%f", lat))
	query.Set("lon", fmt.Sprintf("%f", lon))
	query.Set("units", string(units))
	query.Set("appid", r.APIKey)

	excludeStrs := make([]string, len(exclude))
	for i := range exclude {
		excludeStrs[i] = string(exclude[i])
	}
	query.Set("exclude", strings.Join(excludeStrs, ","))

	reqURL.RawQuery = query.Encode()
	req := (&http.Request{
		Method: http.MethodGet,
		URL:    &reqURL,
		Header: header,
	}).WithContext(ctx)

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body) // ignore failure to read error body
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var typedResp *OnecallResponse
	if err := json.NewDecoder(resp.Body).Decode(&typedResp); err != nil {
		return nil, fmt.Errorf("unmarshaling weather data: %w", err)
	}

	return typedResp, nil
}
