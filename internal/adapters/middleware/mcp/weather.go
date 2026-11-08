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

package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tdrn-org/totem/internal/adapters/integration"
	"github.com/tdrn-org/totem/internal/domain"
	"github.com/tdrn-org/totem/internal/domain/weather"
)

// The weather scopes supported by the getWeather tool. Each scope maps
// 1:1 to a channel of [weather.WeatherService] and defines the
// resolution of the returned points.
const (
	scopeCurrent  = "current"
	scopeUpcoming = "upcoming"
	scope24h      = "24h"
	scope7d       = "7d"
)

// weatherPoint is the transport representation of a single weather point.
// All times are rendered in the time zone of the queried location; the
// instant stays unambiguous since RFC 3339 carries the offset.
type weatherPoint struct {
	Time                time.Time        `json:"time"`
	Temperature         float32          `json:"temperature"`
	ApparentTemperature float32          `json:"apparentTemperature"`
	Humidity            float32          `json:"humidity"`
	Precipitation       float32          `json:"precipitation"`
	WindDirection       float32          `json:"windDirection"`
	WindSpeed           float32          `json:"windSpeed"`
	WindGusts           float32          `json:"windGusts"`
	Condition           weatherCondition `json:"condition"`
	IsDay               bool             `json:"isDay"`
}

// weatherCondition mirrors the WMO condition of a point as structure:
// the numeric WMO code and its stable enum name are the truth, the icon
// is a shortcut for direct rendering. The server provides no localized
// text - the wording is up to the client.
type weatherCondition struct {
	Code int    `json:"code"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// weatherLocation is the resolved location a weather query refers to.
type weatherLocation struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
}

// weatherUnits are the units the provider reported for the queried values.
type weatherUnits struct {
	Temperature   string `json:"temperature"`
	Precipitation string `json:"precipitation"`
	WindSpeed     string `json:"windSpeed"`
}

// weatherDay is the transport representation of a single daily aggregate.
// A daily aggregate is not a measurement at a point in time: it carries
// ranges, sums and maxima for one calendar day. The day itself is
// therefore rendered as a plain date (not as an instant) - a calendar day
// has no unambiguous time zone, and rendering it as a timestamp would
// silently shift it for any consumer in a different time zone.
type weatherDay struct {
	Date                   string           `json:"date"`
	TemperatureMin         float32          `json:"temperatureMin"`
	TemperatureMax         float32          `json:"temperatureMax"`
	ApparentTemperatureMin float32          `json:"apparentTemperatureMin"`
	ApparentTemperatureMax float32          `json:"apparentTemperatureMax"`
	PrecipitationSum       float32          `json:"precipitationSum"`
	WindDirectionDominant  float32          `json:"windDirectionDominant"`
	WindSpeedMax           float32          `json:"windSpeedMax"`
	WindGustsMax           float32          `json:"windGustsMax"`
	Condition              weatherCondition `json:"condition"`
}

// weatherQueryResult bundles the converted points of a single query. A
// scope either yields points (measurements) or days (daily aggregates),
// never both - the scope decides which representation applies.
type weatherQueryResult struct {
	points     []weatherPoint
	days       []weatherDay
	resolution string
	units      *weatherUnits
}

func registerWeatherTools(server *mcp.Server, runtime Runtime) {
	services := filterServices[weather.WeatherService](runtime.Services(), integration.NatureWeather, runtime.Logger())
	if len(services) == 0 {
		return
	}
	registerWeatherAPI(server, services, runtime)
}

func registerWeatherAPI(server *mcp.Server, services []weather.WeatherService, runtime Runtime) {
	// getWeather: current conditions and forecasts of a location
	mcp.AddTool(server, &mcp.Tool{
		Name: "getWeather",
		Description: "Query the weather of a location. The scope selects the channel and with it the resolution of the returned points: " +
			"'current' (one point, now), 'upcoming' (15 minute steps, next 8 hours), " +
			"'24h' (hourly steps, next 24 hours), '7d' (daily aggregates, next 7 days). " +
			"All returned times are rendered in the time zone of the queried location; " +
			"the weather condition is returned as WMO code, stable enum name and icon. " +
			"The '7d' scope returns 'days' (daily aggregates: temperature ranges, sums, maxima and dominant wind direction) " +
			"instead of 'points'; the day is a plain date since a calendar day carries no unambiguous time zone.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"locationID": map[string]any{"type": "string", "description": "Optional location ID to query (e.g. 'Giselher16'). Omit to use the default location."},
				"scope": map[string]any{
					"type":        "string",
					"enum":        []string{scopeCurrent, scopeUpcoming, scope24h, scope7d},
					"default":     scopeCurrent,
					"description": "The weather channel to query.",
				},
			},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		LocationID string `json:"locationID,omitempty"`
		Scope      string `json:"scope,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		scope := input.Scope
		if scope == "" {
			scope = scopeCurrent
		}
		location := runtime.ResolveLocationID(input.LocationID)
		if location.ID == "" {
			return nil, nil, fmt.Errorf("location %q is unknown; pass a location ID from the server configuration (e.g. 'Giselher16')", input.LocationID)
		}
		result, err := queryWeather(ctx, services, scope, location, runtime.Logger())
		if err != nil {
			return nil, nil, err
		}
		text := formatWeatherText(location, scope, result)
		structured := map[string]any{
			"scope":      scope,
			"resolution": result.resolution,
			"location":   newWeatherLocation(location),
			"units":      result.units,
		}
		if result.days != nil {
			structured["days"] = result.days
		} else {
			structured["points"] = result.points
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, structured, nil
	})
}

// queryWeather dispatches the given scope to the weather services. The
// first service answering successfully wins; a service declaring the
// weather nature without being able to answer is not fatal, since other
// services may still be configured.
func queryWeather(ctx context.Context, services []weather.WeatherService, scope string, location *domain.Location, logger *slog.Logger) (*weatherQueryResult, error) {
	var lastErr error
	for _, service := range services {
		result, err := queryWeatherFromService(ctx, service, scope, location)
		if err != nil {
			logger.Warn("weather service query failed", slog.String("scope", scope), slog.Any("err", err))
			lastErr = err
			continue
		}
		return result, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no weather service available")
	}
	return nil, fmt.Errorf("failed to query %q weather for location %q (cause: %w)", scope, location.ID, lastErr)
}

func queryWeatherFromService(ctx context.Context, service weather.WeatherService, scope string, location *domain.Location) (*weatherQueryResult, error) {
	geoLocation := location.GeographicLocation
	switch scope {
	case scopeCurrent:
		current, err := service.QueryCurrentWeather(ctx, geoLocation)
		if err != nil {
			return nil, err
		}
		point := newWeatherPoint(current, location.Timezone)
		return &weatherQueryResult{
			points:     []weatherPoint{point},
			resolution: "point",
			units:      newWeatherUnits(current),
		}, nil
	case scopeUpcoming:
		weathers, err := service.QueryUpcomingWeather(ctx, geoLocation)
		if err != nil {
			return nil, err
		}
		return &weatherQueryResult{
			points:     newWeatherPoints(weathers, location.Timezone),
			resolution: "15min",
			units:      newWeatherUnitsFrom(weathers),
		}, nil
	case scope24h:
		weathers, err := service.Query24hWeather(ctx, geoLocation)
		if err != nil {
			return nil, err
		}
		return &weatherQueryResult{
			points:     newWeatherPoints(weathers, location.Timezone),
			resolution: "hour",
			units:      newWeatherUnitsFrom(weathers),
		}, nil
	case scope7d:
		days, err := service.Query7dWeather(ctx, geoLocation)
		if err != nil {
			return nil, err
		}
		return &weatherQueryResult{
			days:       newWeatherDays(days, location.Timezone),
			resolution: "day",
			units:      newWeatherUnitsFromDaily(days),
		}, nil
	default:
		return nil, fmt.Errorf("unknown scope %q (expected one of %s, %s, %s, %s)",
			scope, scopeCurrent, scopeUpcoming, scope24h, scope7d)
	}
}

func newWeatherPoints(weathers []*weather.Weather, timezone *time.Location) []weatherPoint {
	points := make([]weatherPoint, 0, len(weathers))
	for _, w := range weathers {
		points = append(points, newWeatherPoint(w, timezone))
	}
	return points
}

func newWeatherDays(days []*weather.DailyWeather, timezone *time.Location) []weatherDay {
	result := make([]weatherDay, 0, len(days))
	for _, day := range days {
		result = append(result, newWeatherDay(day, timezone))
	}
	return result
}

// newWeatherDay converts a daily aggregate to its transport
// representation. The day is rendered as a date in the location's time
// zone, since the provider delivers the day boundaries as a plain date.
func newWeatherDay(day *weather.DailyWeather, timezone *time.Location) weatherDay {
	dayTime := day.Time
	if timezone != nil {
		dayTime = dayTime.In(timezone)
	}
	return weatherDay{
		Date:                   dayTime.Format("2006-01-02"),
		TemperatureMin:         day.TemperatureMin,
		TemperatureMax:         day.TemperatureMax,
		ApparentTemperatureMin: day.ApparentTemperatureMin,
		ApparentTemperatureMax: day.ApparentTemperatureMax,
		PrecipitationSum:       day.PrecipitationSum,
		WindDirectionDominant:  day.WindDirectionDominant,
		WindSpeedMax:           day.WindSpeedMax,
		WindGustsMax:           day.WindGustsMax,
		Condition: weatherCondition{
			Code: day.Condition.Code,
			Name: day.Condition.Name,
			Icon: day.Condition.Icon,
		},
	}
}

// newWeatherPoint converts a domain weather point to its transport
// representation. The domain keeps instants in the provider's time zone
// (GMT); rendering in the location's time zone happens here, at the
// presentation boundary.
func newWeatherPoint(w *weather.Weather, timezone *time.Location) weatherPoint {
	pointTime := w.Time
	if timezone != nil {
		pointTime = pointTime.In(timezone)
	}
	return weatherPoint{
		Time:                pointTime,
		Temperature:         w.Temperature,
		ApparentTemperature: w.ApparentTemperature,
		Humidity:            w.Humidity,
		Precipitation:       w.Precipitation,
		WindDirection:       w.WindDirection,
		WindSpeed:           w.WindSpeed,
		WindGusts:           w.WindGusts,
		Condition: weatherCondition{
			Code: w.Condition.Code,
			Name: w.Condition.Name,
			Icon: w.Condition.Icon,
		},
		IsDay: w.IsDay,
	}
}

func newWeatherUnits(w *weather.Weather) *weatherUnits {
	if w == nil {
		return nil
	}
	return &weatherUnits{
		Temperature:   w.TemperatureUnit,
		Precipitation: w.PrecipitationUnit,
		WindSpeed:     w.WindSpeedUnit,
	}
}

func newWeatherUnitsFrom(weathers []*weather.Weather) *weatherUnits {
	if len(weathers) == 0 {
		return nil
	}
	return newWeatherUnits(weathers[0])
}

func newWeatherUnitsFromDaily(days []*weather.DailyWeather) *weatherUnits {
	if len(days) == 0 {
		return nil
	}
	day := days[0]
	return &weatherUnits{
		Temperature:   day.TemperatureUnit,
		Precipitation: day.PrecipitationUnit,
		WindSpeed:     day.WindSpeedUnit,
	}
}

func newWeatherLocation(location *domain.Location) weatherLocation {
	return weatherLocation{
		ID:        location.ID,
		Name:      location.Name,
		Address:   location.Address,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		Timezone:  locationTimezoneName(location),
	}
}

func locationTimezoneName(location *domain.Location) string {
	if location.Timezone == nil {
		return "GMT"
	}
	return location.Timezone.String()
}

func formatWeatherText(location *domain.Location, scope string, result *weatherQueryResult) string {
	text := fmt.Sprintf("Weather for %q (%s) - scope %s, time zone %s:\n",
		location.Name, location.Address, scope, locationTimezoneName(location))
	if result.days != nil {
		return formatWeatherDaysText(text, result.days)
	}
	points := result.points
	if len(points) == 0 {
		return text + "  No weather data available.\n"
	}
	for _, point := range points {
		text += fmt.Sprintf("  %s: %s %.1f (feels like %.1f), humidity %.0f%%, ",
			point.Time.Format("2006-01-02 15:04"), point.Condition.Name,
			point.Temperature, point.ApparentTemperature, point.Humidity)
		text += fmt.Sprintf("wind %.1f from %.0f° (gusts %.1f), precipitation %.1f",
			point.WindSpeed, point.WindDirection, point.WindGusts, point.Precipitation)
		if point.IsDay {
			text += ", day"
		} else {
			text += ", night"
		}
		text += "\n"
	}
	return strings.TrimRight(text, "\n")
}

func formatWeatherDaysText(text string, days []weatherDay) string {
	if len(days) == 0 {
		return text + "  No weather data available.\n"
	}
	for _, day := range days {
		text += fmt.Sprintf("  %s: %s %.1f to %.1f (feels like %.1f to %.1f), ",
			day.Date, day.Condition.Name,
			day.TemperatureMin, day.TemperatureMax,
			day.ApparentTemperatureMin, day.ApparentTemperatureMax)
		text += fmt.Sprintf("precipitation sum %.1f, wind max %.1f from %.0f° (gusts max %.1f)\n",
			day.PrecipitationSum, day.WindSpeedMax, day.WindDirectionDominant, day.WindGustsMax)
	}
	return strings.TrimRight(text, "\n")
}
