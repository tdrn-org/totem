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

package weather

import (
	"context"
	"time"

	"github.com/tdrn-org/totem/internal/domain"
)

type WeatherUnits struct {
	TemperatureUnit   string
	PrecipitationUnit string
	WindSpeedUnit     string
}

type WeatherCondition struct {
	Code int
	Name string
	Icon string
}

type Weather struct {
	WeatherUnits
	Time                time.Time
	Temperature         float32
	ApparentTemperature float32
	Humidity            float32
	WindDirection       float32
	Precipitation       float32
	WindSpeed           float32
	WindGusts           float32
	Condition           WeatherCondition
	IsDay               bool
}

type WeatherService interface {
	QueryCurrentWeather(ctx context.Context, location domain.GeographicLocation) (*Weather, error)
	QueryUpcomingWeather(ctx context.Context, location domain.GeographicLocation) ([]*Weather, error)
	Query24hWeather(ctx context.Context, location domain.GeographicLocation) ([]*Weather, error)
	// Query7dWeather(ctx context.Context, location domain.GeographicLocation) ([]*Weather, error)
}
