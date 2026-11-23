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
	"math"

	"github.com/tdrn-org/go-hue/api"
	"github.com/tdrn-org/totem/internal/domain"
	"github.com/tdrn-org/totem/internal/domain/smarthome"
)

// toLight converts a bridge light response into the domain representation.
//
// The bridge omits attributes a device does not support, so every optional
// attribute has to be nil-checked before it is dereferenced. Dimming is the
// relevant case: devices that cannot be dimmed (e.g. smart plugs) carry no
// dimming attribute at all, and their brightness is reported as 0.
func toLight(lightResponse api.LightGet, location *domain.Location) *smarthome.Light {
	var brightness *api.Brightness
	if lightResponse.Dimming != nil {
		brightness = lightResponse.Dimming.Brightness
	}
	return &smarthome.Light{
		ID:         lightResponse.Id,
		Name:       lightResponse.Metadata.Name,
		On:         lightResponse.On.On,
		Brightness: brightnessValue(brightness),
		Location:   location,
	}
}

// brightnessValue converts the bridge brightness attribute into the domain
// brightness. The bridge quantizes its brightness onto a 253 step grid
// (level = round(1 + percent * 2.53), percent = (level - 1) / 253 * 100), so a
// light asked to dim to 80% reports back 79.84 and one asked for 56% reports
// 56.13. The value therefore has to be rounded: truncating would report a
// brightness up to one full point below what was set, which is exactly the
// drift this conversion is supposed to avoid.
func brightnessValue(response *api.Brightness) uint8 {
	if response == nil {
		return 0
	}
	value := uint8(math.Round(float64(*response)))
	if value > 100 {
		value = 100
	}
	return value
}

// brightnessPayload converts a domain brightness (0-100) into the bridge's
// brightness attribute, which is a percentage as well. The bridge does not
// accept 0, it falls back to the lowest possible brightness, so 0 is sent as 1.
func brightnessPayload(brightness uint8) api.Brightness {
	value := brightness
	if value > 100 {
		value = 100
	}
	if value < 1 {
		value = 1
	}
	return api.Brightness(value)
}
