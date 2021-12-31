package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alecholmes/weathergrid/pkg/model"
	"github.com/alecholmes/weathergrid/pkg/openweathermap"
)

type snapshotWriter func(context.Context, *model.WeatherSnapshot) error

func NewWeatherSyncer(
	weatherClient openweathermap.Client,
	writer snapshotWriter,
	logger *log.Logger,
) *WeatherSyncer {
	return &WeatherSyncer{
		weatherClient: weatherClient,
		writer:        writer,
		logger:        logger,
	}
}

type WeatherSyncer struct {
	weatherClient openweathermap.Client
	writer        snapshotWriter
	logger        *log.Logger
}

func (w *WeatherSyncer) Sync(ctx context.Context, groups []*locationGroup) error {
	type latLong [2]float64
	respCache := make(map[latLong]*openweathermap.OnecallResponse)

	out := &model.WeatherSnapshot{
		Version: "1",
	}
	for _, g := range groups {
		w.logger.Printf("processing group. group=%s", g.Name)
		group := &model.LocationGroup{Name: g.Name, Slug: g.Slug}
		out.Groups = append(out.Groups, group)

		for _, l := range g.Locations {
			w.logger.Printf("processing location. group=%s location=%s", g.Name, l.Name)
			loc := &location{
				Name: l.Name,
				Slug: l.Slug,
				Lat:  l.Lat,
				Lon:  l.Lon,
			}

			resp, ok := respCache[latLong{l.Lat, l.Lon}]
			if !ok {
				w.logger.Printf("getting weather. group=%s location=%s", group.Name, loc.Name)
				var err error
				resp, err = w.weatherClient.OneCall(ctx, loc.Lat, loc.Lon,
					[]openweathermap.ExcludeType{
						openweathermap.ExcludeMinutely,
						openweathermap.ExcludeDaily,
						openweathermap.ExcludeAlerts,
					},
					openweathermap.UnitsImperial)
				if err != nil {
					return fmt.Errorf("getting weather for %s: %w", loc.Name, err)
				}

				respCache[latLong{l.Lat, l.Lon}] = resp
			}

			group.Locations = append(group.Locations, responseToModel(loc, resp))
		}
	}

	w.logger.Printf("writing snapshot")
	if err := w.writer(ctx, out); err != nil {
		return fmt.Errorf("writing weather: %w", err)
	}

	w.logger.Printf("done")

	return nil
}

func responseToModel(loc *location, resp *openweathermap.OnecallResponse) *model.LocationWeather {
	tz := time.FixedZone(resp.Timezone, resp.TimezoneOffset)

	forecasts := make([]*model.Weather, len(resp.Hourly))
	for i, w := range resp.Hourly {
		forecasts[i] = weatherToModel(w, tz)
	}

	return &model.LocationWeather{
		Name:      loc.Name,
		Slug:      loc.Slug,
		Lat:       resp.Lat,
		Lon:       resp.Lon,
		Latest:    weatherToModel(resp.Current, tz),
		Forecasts: forecasts,
	}
}

func weatherToModel(w *openweathermap.Weather, tz *time.Location) *model.Weather {
	details := make([]*model.WeatherDetails, len(w.Details))
	for i, d := range w.Details {
		details[i] = &model.WeatherDetails{
			Summary:     d.Main,
			Description: d.Description,
		}
	}

	return &model.Weather{
		Timestamp:            time.Unix(w.Timestamp, 0).In(tz),
		Temp:                 w.Temp,
		TempFeelsLike:        w.TempFeelsLike,
		HumidityPercent:      w.HumidityPercent,
		CloudinessPercent:    w.CloudinessPercent,
		WindSpeed:            w.WindSpeed,
		WindGustSpeed:        w.WindGustSpeed,
		WindDegree:           w.WindDegree,
		PrecipitationPercent: w.PrecipitationPercent,
		Details:              details,
	}
}
