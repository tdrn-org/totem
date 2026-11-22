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

package huebridge

import (
	"encoding/json"
	"testing"

	"github.com/tdrn-org/go-hue/api"
	"github.com/tdrn-org/totem/internal/domain"
)

func lightFromJSON(t *testing.T, raw string) api.LightGet {
	t.Helper()
	var light api.LightGet
	if err := json.Unmarshal([]byte(raw), &light); err != nil {
		t.Fatalf("cannot unmarshal light response: %v", err)
	}
	return light
}

// A smart plug cannot be dimmed, so the bridge omits the dimming attribute
// entirely. Converting such a light must not dereference the missing attribute
// (regression: nil pointer panic which killed the complete MCP server).
func TestToLightWithoutDimming(t *testing.T) {
	light := lightFromJSON(t, `{
		"id": "d6edb842-8e03-4a53-b047-e8ce88615458",
		"metadata": {"name": "Steckdose Arbeitszimmer", "archetype": "plug"},
		"on": {"on": true}
	}`)
	location := &domain.Location{ID: "Giselher16"}
	converted := toLight(light, location)
	if converted == nil {
		t.Fatal("expected a light, got nil")
	}
	if converted.ID != "d6edb842-8e03-4a53-b047-e8ce88615458" {
		t.Errorf("unexpected ID: %q", converted.ID)
	}
	if converted.Name != "Steckdose Arbeitszimmer" {
		t.Errorf("unexpected name: %q", converted.Name)
	}
	if !converted.On {
		t.Error("expected plug to be switched on")
	}
	if converted.Brightness != 0 {
		t.Errorf("expected brightness 0 for a non-dimmable device, got %d", converted.Brightness)
	}
	if converted.Location != location {
		t.Error("expected the given location to be preserved")
	}
}

// An empty dimming attribute is possible as well: the attribute is present, but
// the bridge did not report a brightness value.
func TestToLightWithEmptyDimming(t *testing.T) {
	light := lightFromJSON(t, `{
		"id": "1",
		"metadata": {"name": "Flur 1"},
		"on": {"on": false},
		"dimming": {}
	}`)
	converted := toLight(light, nil)
	if converted.Brightness != 0 {
		t.Errorf("expected brightness 0, got %d", converted.Brightness)
	}
	if converted.On {
		t.Error("expected light to be switched off")
	}
}

func TestToLightWithDimming(t *testing.T) {
	light := lightFromJSON(t, `{
		"id": "2",
		"metadata": {"name": "Arbeitszimmer 1"},
		"on": {"on": true},
		"dimming": {"brightness": 56.52, "min_dim_level": 0.2}
	}`)
	converted := toLight(light, nil)
	if converted.Brightness != 56 {
		t.Errorf("expected brightness 56, got %d", converted.Brightness)
	}
}

func TestBrightnessValue(t *testing.T) {
	tests := []struct {
		name     string
		response *api.Brightness
		expected uint8
	}{
		{name: "nil", response: nil, expected: 0},
		{name: "zero", response: brightness(0.0), expected: 0},
		{name: "bridge default", response: brightness(56.52), expected: 56},
		{name: "full", response: brightness(100.0), expected: 100},
		{name: "out of range", response: brightness(150.0), expected: 100},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := brightnessValue(test.response)
			if actual != test.expected {
				t.Errorf("expected %d, got %d", test.expected, actual)
			}
		})
	}
}

// The bridge brightness attribute is a percentage, not a fraction — sending a
// fraction would dim a light down to (almost) nothing.
func TestBrightnessPayload(t *testing.T) {
	tests := []struct {
		name       string
		brightness uint8
		expected   api.Brightness
	}{
		{name: "zero becomes lowest brightness", brightness: 0, expected: 1},
		{name: "percentage is passed through", brightness: 56, expected: 56},
		{name: "full", brightness: 100, expected: 100},
		{name: "out of range", brightness: 150, expected: 100},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := brightnessPayload(test.brightness)
			if actual != test.expected {
				t.Errorf("expected %v, got %v", test.expected, actual)
			}
		})
	}
}

func brightness(value float32) *api.Brightness {
	brightness := api.Brightness(value)
	return &brightness
}
