/*
 * Copyright 2026 Holger de Carne
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package openmeteo

import (
	"context"
	"fmt"
	"slices"
	"time"

	openmeteoapi "github.com/tdrn-org/go-open-meteo/api"
	"github.com/tdrn-org/totem/internal/domain"
	"github.com/tdrn-org/totem/internal/domain/weather"
)

func (s *openmeteoService) QueryCurrentWeather(ctx context.Context, location domain.GeographicLocation) (*weather.Weather, error) {
	params := newForecastParams(s.config, location).
		Current().
		Build()
	response, err := s.apiClient.GetV1Forecast(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query current weather (cause: %w)", err)
	}
	weather, err := currentResponseToCurrentWeather((*openmeteoapi.GetV1Forecast200JSONResponse)(response.JSON200))
	if err != nil {
		return nil, fmt.Errorf("failed to convert current weather response (cause: %w)", err)
	}
	return weather, nil
}

func (s *openmeteoService) QueryUpcomingWeather(ctx context.Context, location domain.GeographicLocation) ([]*weather.Weather, error) {
	params := newForecastParams(s.config, location).
		Minutely15(2, 8).
		Build()
	response, err := s.apiClient.GetV1Forecast(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query upcoming weather (cause: %w)", err)
	}
	weathers, err := minutely15ResponseToWeathers((*openmeteoapi.GetV1Forecast200JSONResponse)(response.JSON200))
	if err != nil {
		return nil, fmt.Errorf("failed to convert upcoming weather response (cause: %w)", err)
	}
	return sortAndFilterWeathers(weathers, time.Second, 8*time.Hour), nil
}

func (s *openmeteoService) Query24hWeather(ctx context.Context, location domain.GeographicLocation) ([]*weather.Weather, error) {
	params := newForecastParams(s.config, location).
		Hourly(2, 24).
		Build()
	response, err := s.apiClient.GetV1Forecast(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query 24h weather (cause: %w)", err)
	}
	weathers, err := hourlyResponseToWeathers((*openmeteoapi.GetV1Forecast200JSONResponse)(response.JSON200))
	if err != nil {
		return nil, fmt.Errorf("failed to convert 24h weather response (cause: %w)", err)
	}
	return sortAndFilterWeathers(weathers, time.Hour, 24*time.Hour), nil
}

func (s *openmeteoService) Query7dWeather(ctx context.Context, location domain.GeographicLocation) ([]*weather.AggregatedWeather, error) {
	params := newForecastParams(s.config, location).
		Daily(8).
		Build()
	response, err := s.apiClient.GetV1Forecast(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query 7d weather (cause: %w)", err)
	}
	weathers, err := dailyResponseToDailyWeathers((*openmeteoapi.GetV1Forecast200JSONResponse)(response.JSON200))
	if err != nil {
		return nil, fmt.Errorf("failed to convert 7d weather response (cause: %w)", err)
	}
	return sortAndFilterDailyWeathers(weathers, time.Hour, 7*24*time.Hour), nil
}

type forecastParamsBuilder struct {
	params *openmeteoapi.GetV1ForecastParams
}

func newForecastParams(config Config, location domain.GeographicLocation) *forecastParamsBuilder {
	temperaturUnit := openmeteoapi.GetV1ForecastParamsTemperatureUnit(config.TemperatureUnit())
	precipitationUnit := openmeteoapi.GetV1ForecastParamsPrecipitationUnit(config.PrecipitationUnit())
	windSpeedUnit := openmeteoapi.GetV1ForecastParamsWindSpeedUnit(config.WindSpeedUnit())
	return &forecastParamsBuilder{
		params: &openmeteoapi.GetV1ForecastParams{
			Latitude:          fmt.Sprintf("%.4f", location.Latitude),
			Longitude:         fmt.Sprintf("%.4f", location.Longitude),
			TemperatureUnit:   &temperaturUnit,
			PrecipitationUnit: &precipitationUnit,
			WindSpeedUnit:     &windSpeedUnit,
		},
	}
}

var paramsCurrent []openmeteoapi.GetV1ForecastParamsCurrent = []openmeteoapi.GetV1ForecastParamsCurrent{
	openmeteoapi.GetV1ForecastParamsCurrentTemperature2m,
	openmeteoapi.GetV1ForecastParamsCurrentApparentTemperature,
	openmeteoapi.GetV1ForecastParamsCurrentRelativeHumidity2m,
	openmeteoapi.GetV1ForecastParamsCurrentCloudCover,
	openmeteoapi.GetV1ForecastParamsCurrentWindDirection10m,
	openmeteoapi.GetV1ForecastParamsCurrentPrecipitation,
	openmeteoapi.GetV1ForecastParamsCurrentWindSpeed10m,
	openmeteoapi.GetV1ForecastParamsCurrentWindGusts10m,
	openmeteoapi.GetV1ForecastParamsCurrentWeatherCode,
	openmeteoapi.GetV1ForecastParamsCurrentIsDay,
}

func (b *forecastParamsBuilder) Current() *forecastParamsBuilder {
	b.params.Current = &paramsCurrent
	return b
}

var paramsMinutely15 []openmeteoapi.GetV1ForecastParamsMinutely15 = []openmeteoapi.GetV1ForecastParamsMinutely15{
	openmeteoapi.GetV1ForecastParamsMinutely15Temperature2m,
	openmeteoapi.GetV1ForecastParamsMinutely15ApparentTemperature,
	openmeteoapi.GetV1ForecastParamsMinutely15RelativeHumidity2m,
	openmeteoapi.GetV1ForecastParamsMinutely15WindDirection10m,
	openmeteoapi.GetV1ForecastParamsMinutely15Precipitation,
	openmeteoapi.GetV1ForecastParamsMinutely15WindSpeed10m,
	openmeteoapi.GetV1ForecastParamsMinutely15WindGusts10m,
	openmeteoapi.GetV1ForecastParamsMinutely15WeatherCode,
	openmeteoapi.GetV1ForecastParamsMinutely15IsDay,
}

func (b *forecastParamsBuilder) Minutely15(days, hours int) *forecastParamsBuilder {
	b.params.Minutely15 = &paramsMinutely15
	b.params.ForecastDays = &days
	b.params.ForecastHours = &hours
	return b
}

var paramsHourly []openmeteoapi.GetV1ForecastParamsHourly = []openmeteoapi.GetV1ForecastParamsHourly{
	openmeteoapi.GetV1ForecastParamsHourlyTemperature2m,
	openmeteoapi.GetV1ForecastParamsHourlyApparentTemperature,
	openmeteoapi.GetV1ForecastParamsHourlyRelativeHumidity2m,
	openmeteoapi.GetV1ForecastParamsHourlyWindDirection10m,
	openmeteoapi.GetV1ForecastParamsHourlyPrecipitation,
	openmeteoapi.GetV1ForecastParamsHourlyWindSpeed10m,
	openmeteoapi.GetV1ForecastParamsHourlyWindGusts10m,
	openmeteoapi.GetV1ForecastParamsHourlyWeatherCode,
	openmeteoapi.GetV1ForecastParamsHourlyIsDay,
}

func (b *forecastParamsBuilder) Hourly(days, hours int) *forecastParamsBuilder {
	b.params.Hourly = &paramsHourly
	b.params.ForecastDays = &days
	b.params.ForecastHours = &hours
	return b
}

var paramsDaily []openmeteoapi.GetV1ForecastParamsDaily = []openmeteoapi.GetV1ForecastParamsDaily{
	openmeteoapi.GetV1ForecastParamsDailyTemperature2mMin,
	openmeteoapi.GetV1ForecastParamsDailyTemperature2mMax,
	openmeteoapi.GetV1ForecastParamsDailyApparentTemperatureMin,
	openmeteoapi.GetV1ForecastParamsDailyApparentTemperatureMax,
	openmeteoapi.GetV1ForecastParamsDailyWindDirection10mDominant,
	openmeteoapi.GetV1ForecastParamsDailyPrecipitationSum,
	openmeteoapi.GetV1ForecastParamsDailyWindSpeed10mMax,
	openmeteoapi.GetV1ForecastParamsDailyWindGusts10mMax,
	openmeteoapi.GetV1ForecastParamsDailyWeatherCode,
}

func (b *forecastParamsBuilder) Daily(days int) *forecastParamsBuilder {
	b.params.Daily = &paramsDaily
	b.params.ForecastDays = &days
	return b
}

func (b *forecastParamsBuilder) Build() *openmeteoapi.GetV1ForecastParams {
	return b.params
}

func sortAndFilterWeathers(weathers []*weather.Weather, truncate, timerange time.Duration) []*weather.Weather {
	from := time.Now().Truncate(truncate)
	to := from.Add(timerange)
	result := make([]*weather.Weather, 0, len(weathers))
	for _, weather := range weathers {
		if weather.Time.Compare(from) >= 0 && weather.Time.Compare(to) < 0 {
			result = append(result, weather)
		}
	}
	slices.SortFunc(result, func(a, b *weather.Weather) int {
		return a.Time.Compare(b.Time)
	})
	return result
}

func sortAndFilterDailyWeathers(weathers []*weather.AggregatedWeather, truncate, timerange time.Duration) []*weather.AggregatedWeather {
	from := time.Now().Truncate(truncate)
	to := from.Add(timerange)
	result := make([]*weather.AggregatedWeather, 0, len(weathers))
	for _, weather := range weathers {
		if weather.Time.Compare(from) >= 0 && weather.Time.Compare(to) < 0 {
			result = append(result, weather)
		}
	}
	slices.SortFunc(result, func(a, b *weather.AggregatedWeather) int {
		return a.Time.Compare(b.Time)
	})
	return result
}
