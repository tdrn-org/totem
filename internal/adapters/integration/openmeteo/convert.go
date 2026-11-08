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
	"fmt"

	openmeteo "github.com/tdrn-org/go-open-meteo"
	openmeteoapi "github.com/tdrn-org/go-open-meteo/api"
	"github.com/tdrn-org/totem/internal/domain/weather"
)

func currentResponseToCurrentWeather(response *openmeteoapi.GetV1Forecast200JSONResponse) (*weather.Weather, error) {
	if response.CurrentUnits == nil || response.CurrentUnits.Temperature2m == nil || response.CurrentUnits.Precipitation == nil || response.CurrentUnits.WindSpeed10m == nil {
		return nil, fmt.Errorf("missing current units")
	}
	if response.Current == nil || response.Current.Time == nil || response.Timezone == nil {
		return nil, fmt.Errorf("missing current time/timezone")
	}
	weatherTime, err := openmeteo.GetTime(response.Current.Time, *response.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid current time/timezone (cause: %w)", err)
	}
	if response.Current.Temperature2m == nil {
		return nil, fmt.Errorf("missing current temperature (2m)")
	}
	if response.Current.ApparentTemperature == nil {
		return nil, fmt.Errorf("missing current apparent temperature")
	}
	if response.Current.RelativeHumidity2m == nil {
		return nil, fmt.Errorf("missing current relative humidity")
	}
	if response.Current.WindDirection10m == nil {
		return nil, fmt.Errorf("missing current wind direction (10m)")
	}
	if response.Current.Precipitation == nil {
		return nil, fmt.Errorf("missing current precipitation")
	}
	if response.Current.WindSpeed10m == nil {
		return nil, fmt.Errorf("missing current wind speed (10m)")
	}
	if response.Current.WindGusts10m == nil {
		return nil, fmt.Errorf("missing current wind gusts (10m)")
	}
	if response.Current.WeatherCode == nil {
		return nil, fmt.Errorf("missing current weather code")
	}
	weatherCode := openmeteo.WMOCode(*response.Current.WeatherCode)
	if response.Current.IsDay == nil {
		return nil, fmt.Errorf("missing current day flag")
	}
	weather := &weather.Weather{
		TemperatureUnit:     *response.CurrentUnits.Temperature2m,
		PrecipitationUnit:   *response.CurrentUnits.Precipitation,
		WindSpeedUnit:       *response.CurrentUnits.WindSpeed10m,
		Time:                weatherTime,
		Temperature:         *response.Current.Temperature2m,
		ApparentTemperature: *response.Current.ApparentTemperature,
		Humidity:            *response.Current.RelativeHumidity2m,
		WindDirection:       *response.Current.WindDirection10m,
		Precipitation:       *response.Current.Precipitation,
		WindSpeed:           *response.Current.WindSpeed10m,
		WindGusts:           *response.Current.WindGusts10m,
		Condition: weather.WeatherCondition{
			Code: int(weatherCode),
			Name: weatherCode.Code(),
			Icon: weatherCode.Icon(),
		},
		IsDay: (*response.Current.IsDay) != 0,
	}
	return weather, nil
}

func minutely15ResponseToWeathers(response *openmeteoapi.GetV1Forecast200JSONResponse) ([]*weather.Weather, error) {
	if response.Minutely15Units == nil || response.Minutely15Units.Temperature2m == nil || response.Minutely15Units.Precipitation == nil || response.Minutely15Units.WindSpeed10m == nil {
		return nil, fmt.Errorf("missing minutely15 units")
	}
	if response.Minutely15 == nil || response.Minutely15.Time == nil || response.Timezone == nil {
		return nil, fmt.Errorf("missing minutely15 time/timezone")
	}
	weatherCount := len(*response.Minutely15.Time)
	if weatherCount == 0 {
		return nil, fmt.Errorf("missing or incomplelte minutely15 response")
	}
	weathers := make([]*weather.Weather, 0, weatherCount)
	for i := range weatherCount {
		weather, err := minutely15ResponseToWeather(response, i)
		if err != nil {
			return nil, err
		}
		weathers = append(weathers, weather)
	}
	return weathers, nil
}

func minutely15ResponseToWeather(response *openmeteoapi.GetV1Forecast200JSONResponse, i int) (*weather.Weather, error) {
	weatherTime, err := openmeteo.GetTime(&(*response.Minutely15.Time)[i], *response.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid minutely15 time/timezone (cause: %w)", err)
	}
	if response.Minutely15.Temperature2m == nil || len(*response.Minutely15.Temperature2m) <= i {
		return nil, fmt.Errorf("missing minutely15 temperature (2m)")
	}
	if response.Minutely15.ApparentTemperature == nil || len(*response.Minutely15.ApparentTemperature) <= i {
		return nil, fmt.Errorf("missing minutely15 apparent temperature")
	}
	if response.Minutely15.RelativeHumidity2m == nil || len(*response.Minutely15.RelativeHumidity2m) <= i {
		return nil, fmt.Errorf("missing minutely15 relative humidity")
	}
	if response.Minutely15.WindDirection10m == nil || len(*response.Minutely15.WindDirection10m) <= i {
		return nil, fmt.Errorf("missing minutely15 wind direction (10m)")
	}
	if response.Minutely15.Precipitation == nil || len(*response.Minutely15.Precipitation) <= i {
		return nil, fmt.Errorf("missing minutely15 precipitation")
	}
	if response.Minutely15.WindSpeed10m == nil || len(*response.Minutely15.WindSpeed10m) <= i {
		return nil, fmt.Errorf("missing minutely15 wind speed (10m)")
	}
	if response.Minutely15.WindGusts10m == nil || len(*response.Minutely15.WindGusts10m) <= i {
		return nil, fmt.Errorf("missing minutely15 wind gusts (10m)")
	}
	if response.Minutely15.WeatherCode == nil || len(*response.Minutely15.WeatherCode) <= i {
		return nil, fmt.Errorf("missing minutely15 weather code")
	}
	weatherCode := openmeteo.WMOCode((*response.Minutely15.WeatherCode)[i])
	if response.Minutely15.IsDay == nil || len(*response.Minutely15.IsDay) <= i {
		return nil, fmt.Errorf("missing minutely15 day flag")
	}
	weather := &weather.Weather{
		TemperatureUnit:     *response.Minutely15Units.Temperature2m,
		PrecipitationUnit:   *response.Minutely15Units.Precipitation,
		WindSpeedUnit:       *response.Minutely15Units.WindSpeed10m,
		Time:                weatherTime,
		Temperature:         (*response.Minutely15.Temperature2m)[i],
		ApparentTemperature: (*response.Minutely15.ApparentTemperature)[i],
		Humidity:            (*response.Minutely15.RelativeHumidity2m)[i],
		WindDirection:       (*response.Minutely15.WindDirection10m)[i],
		Precipitation:       (*response.Minutely15.Precipitation)[i],
		WindSpeed:           (*response.Minutely15.WindSpeed10m)[i],
		WindGusts:           (*response.Minutely15.WindGusts10m)[i],
		Condition: weather.WeatherCondition{
			Code: int(weatherCode),
			Name: weatherCode.Code(),
			Icon: weatherCode.Icon(),
		},
		IsDay: (*response.Minutely15.IsDay)[i] != 0,
	}
	return weather, nil
}

func hourlyResponseToWeathers(response *openmeteoapi.GetV1Forecast200JSONResponse) ([]*weather.Weather, error) {
	if response.HourlyUnits == nil || response.HourlyUnits.Temperature2m == nil || response.HourlyUnits.Precipitation == nil || response.HourlyUnits.WindSpeed10m == nil {
		return nil, fmt.Errorf("missing hourly units")
	}
	if response.Hourly == nil || response.Hourly.Time == nil || response.Timezone == nil {
		return nil, fmt.Errorf("missing hourly time/timezone")
	}
	weatherCount := len(*response.Hourly.Time)
	if weatherCount == 0 {
		return nil, fmt.Errorf("missing or incomplelte hourly response")
	}
	weathers := make([]*weather.Weather, 0, weatherCount)
	for i := range weatherCount {
		weather, err := hourlyResponseToWeather(response, i)
		if err != nil {
			return nil, err
		}
		weathers = append(weathers, weather)
	}
	return weathers, nil
}

func hourlyResponseToWeather(response *openmeteoapi.GetV1Forecast200JSONResponse, i int) (*weather.Weather, error) {
	weatherTime, err := openmeteo.GetTime(&(*response.Hourly.Time)[i], *response.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid hourly time/timezone (cause: %w)", err)
	}
	if response.Hourly.Temperature2m == nil || len(*response.Hourly.Temperature2m) <= i {
		return nil, fmt.Errorf("missing hourly temperature (2m)")
	}
	if response.Hourly.ApparentTemperature == nil || len(*response.Hourly.ApparentTemperature) <= i {
		return nil, fmt.Errorf("missing hourly apparent temperature")
	}
	if response.Hourly.WindDirection10m == nil || len(*response.Hourly.WindDirection10m) <= i {
		return nil, fmt.Errorf("missing hourly wind direction (10m)")
	}
	if response.Hourly.Precipitation == nil || len(*response.Hourly.Precipitation) <= i {
		return nil, fmt.Errorf("missing hourly precipitation")
	}
	if response.Hourly.WindSpeed10m == nil || len(*response.Hourly.WindSpeed10m) <= i {
		return nil, fmt.Errorf("missing hourly wind speed (10m)")
	}
	if response.Hourly.WindGusts10m == nil || len(*response.Hourly.WindGusts10m) <= i {
		return nil, fmt.Errorf("missing hourly wind gusts (10m)")
	}
	if response.Hourly.WeatherCode == nil || len(*response.Hourly.WeatherCode) <= i {
		return nil, fmt.Errorf("missing hourly weather code")
	}
	weatherCode := openmeteo.WMOCode((*response.Hourly.WeatherCode)[i])
	if response.Hourly.IsDay == nil || len(*response.Hourly.IsDay) <= i {
		return nil, fmt.Errorf("missing hourly day flag")
	}
	weather := &weather.Weather{
		TemperatureUnit:     *response.HourlyUnits.Temperature2m,
		PrecipitationUnit:   *response.HourlyUnits.Precipitation,
		WindSpeedUnit:       *response.HourlyUnits.WindSpeed10m,
		Time:                weatherTime,
		Temperature:         (*response.Hourly.Temperature2m)[i],
		ApparentTemperature: (*response.Hourly.ApparentTemperature)[i],
		Humidity:            (*response.Hourly.RelativeHumidity2m)[i],
		WindDirection:       (*response.Hourly.WindDirection10m)[i],
		Precipitation:       (*response.Hourly.Precipitation)[i],
		WindSpeed:           (*response.Hourly.WindSpeed10m)[i],
		WindGusts:           (*response.Hourly.WindGusts10m)[i],
		Condition: weather.WeatherCondition{
			Code: int(weatherCode),
			Name: weatherCode.Code(),
			Icon: weatherCode.Icon(),
		},
		IsDay: (*response.Hourly.IsDay)[i] != 0,
	}
	return weather, nil
}
