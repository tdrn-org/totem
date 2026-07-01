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
	"time"
)

type LocationConfig struct {
	ID        string       `toml:"id"`
	Name      string       `toml:"name"`
	Address   string       `toml:"address"`
	Latitude  float64      `toml:"latitude"`
	Longitude float64      `toml:"longitude"`
	Timezone  TimezoneSpec `toml:"timezone"`
}

type TimezoneSpec struct {
	*time.Location
}

func (spec *TimezoneSpec) Value() string {
	if spec.Location == nil {
		return ""
	}
	return spec.String()
}

func (spec *TimezoneSpec) MarshalTOML() ([]byte, error) {
	return []byte(`"` + spec.Value() + `"`), nil
}

func (spec *TimezoneSpec) UnmarshalTOML(value any) error {
	locationString, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected location type %v", value)
	}
	location := time.Local
	if locationString != "" {
		parsedLocation, err := time.LoadLocation(locationString)
		if err != nil {
			return fmt.Errorf("invalid location: '%s' (cause: %w)", locationString, err)
		}
		location = parsedLocation
	}
	spec.Location = location
	return nil
}
