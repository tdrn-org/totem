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

package domain

// Location represents a physical location where Totem operates.
// Usually a location is assigned to a specific device (Hue bridge,
// FRITZ!Box, etc.) which operating in exactly this location.
type Location struct {
	// ID is the stable, unique identifier for this location (e.g. "Home").
	ID string
	// Name is the human-readable name of this location.
	Name string
	// Address is the postal address of this location.
	Address string
	// Latitude is the geographic latitude of this location.
	Latitude float64
	// Longitude is the geographic longitude of this location.
	Longitude float64
}

func (l *Location) Equal(other *Location) bool {
	return l.ID == other.ID
}
