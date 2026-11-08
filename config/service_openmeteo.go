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

package config

import (
	"fmt"
	"log/slog"

	"github.com/tdrn-org/go-config-toml"
	openmeteoapi "github.com/tdrn-org/go-open-meteo/api"
)

type OpenMeteoConfig struct {
	Enabled           bool              `toml:"enabled"`
	APIURL            config.URLSpec    `toml:"api_url"`
	TemperatureUnit   TemperatureUnit   `toml:"temperature_unit"`
	PrecipitationUnit PrecipitationUnit `toml:"precipitation_unit"`
	WindSpeedUnit     WindSpeedUnit     `toml:"wind_speed_unit"`
}

type TemperatureUnit string

const (
	TemperatureUnitCelsius    TemperatureUnit = TemperatureUnit(openmeteoapi.Celsius)
	TemperatureUnitFahrenheit TemperatureUnit = TemperatureUnit(openmeteoapi.Fahrenheit)
)

var knownTemperatureUnits map[string]TemperatureUnit = map[string]TemperatureUnit{
	string(TemperatureUnitCelsius):    TemperatureUnitCelsius,
	string(TemperatureUnitFahrenheit): TemperatureUnitFahrenheit,
}

func (u TemperatureUnit) Value() string {
	for value, unit := range knownTemperatureUnits {
		if u == unit {
			return value
		}
	}
	slog.Warn("unexpected temperature unit", slog.Any("u", u))
	return ""
}

func (u *TemperatureUnit) MarshalTOML() ([]byte, error) {
	return []byte(`"` + u.Value() + `"`), nil
}

func (u *TemperatureUnit) UnmarshalTOML(value any) error {
	unitString, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected temperature unit type %v", value)
	}
	unit, ok := knownTemperatureUnits[unitString]
	if !ok {
		return fmt.Errorf("unknown temperature unit: %q", unitString)
	}
	*u = unit
	return nil
}

type PrecipitationUnit string

const (
	PrecipitationUnitInch PrecipitationUnit = PrecipitationUnit(openmeteoapi.Inch)
	PrecipitationUnitMm   PrecipitationUnit = PrecipitationUnit(openmeteoapi.Mm)
)

var knownPrecipitationUnits map[string]PrecipitationUnit = map[string]PrecipitationUnit{
	string(PrecipitationUnitInch): PrecipitationUnitInch,
	string(PrecipitationUnitMm):   PrecipitationUnitMm,
}

func (u PrecipitationUnit) Value() string {
	for value, unit := range knownPrecipitationUnits {
		if u == unit {
			return value
		}
	}
	slog.Warn("unexpected wind speed unit", slog.Any("u", u))
	return ""
}

func (u *PrecipitationUnit) MarshalTOML() ([]byte, error) {
	return []byte(`"` + u.Value() + `"`), nil
}

func (u *PrecipitationUnit) UnmarshalTOML(value any) error {
	unitString, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected wind speed unit type %v", value)
	}
	unit, ok := knownPrecipitationUnits[unitString]
	if !ok {
		return fmt.Errorf("unknown wind speed unit: %q", unitString)
	}
	*u = unit
	return nil
}

type WindSpeedUnit string

const (
	WindSpeedUnitKmh WindSpeedUnit = WindSpeedUnit(openmeteoapi.Kmh)
	WindSpeedUnitKn  WindSpeedUnit = WindSpeedUnit(openmeteoapi.Kn)
	WindSpeedUnitMph WindSpeedUnit = WindSpeedUnit(openmeteoapi.Mph)
	WindSpeedUnitMs  WindSpeedUnit = WindSpeedUnit(openmeteoapi.Ms)
)

var knownWindSpeedUnits map[string]WindSpeedUnit = map[string]WindSpeedUnit{
	string(WindSpeedUnitKmh): WindSpeedUnitKmh,
	string(WindSpeedUnitKn):  WindSpeedUnitKn,
	string(WindSpeedUnitMph): WindSpeedUnitMph,
	string(WindSpeedUnitMs):  WindSpeedUnitMs,
}

func (u WindSpeedUnit) Value() string {
	for value, unit := range knownWindSpeedUnits {
		if u == unit {
			return value
		}
	}
	slog.Warn("unexpected wind speed unit", slog.Any("u", u))
	return ""
}

func (u *WindSpeedUnit) MarshalTOML() ([]byte, error) {
	return []byte(`"` + u.Value() + `"`), nil
}

func (u *WindSpeedUnit) UnmarshalTOML(value any) error {
	unitString, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected wind speed unit type %v", value)
	}
	unit, ok := knownWindSpeedUnits[unitString]
	if !ok {
		return fmt.Errorf("unknown wind speed unit: %q", unitString)
	}
	*u = unit
	return nil
}
