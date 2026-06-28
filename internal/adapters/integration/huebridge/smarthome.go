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
	"context"
	"fmt"

	"github.com/tdrn-org/totem/internal/domain/smarthome"
)

func (s *huebridgeService) ListLights(ctx context.Context) ([]*smarthome.Light, error) {
	bridgeClient, err := s.resolveBridgeClient()
	if err != nil {
		return nil, err
	}
	response, err := bridgeClient.GetLights(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get lights from Hue Bridge (cause: %w)", err)
	}
	lights := make([]*smarthome.Light, 0, len(response.JSON200.Data))
	for _, lightResponse := range response.JSON200.Data {
		light := &smarthome.Light{
			ID:         lightResponse.Id,
			Name:       lightResponse.Metadata.Name,
			On:         lightResponse.On.On,
			Brightness: brightnessValue(lightResponse.Dimming.Brightness),
			Location:   s.host.ResolveLocationID(s.config.LocationID()),
		}
		lights = append(lights, light)
	}
	return lights, nil
}

func (s *huebridgeService) GetLight(ctx context.Context, id string) (*smarthome.Light, error) {
	bridgeClient, err := s.resolveBridgeClient()
	if err != nil {
		return nil, err
	}
	response, err := bridgeClient.GetLight(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get light from Hue Bridge (cause: %w)", err)
	}
	if len(response.JSON200.Data) == 0 {
		return nil, fmt.Errorf("no light data from Hue Bridge")
	}
	lightResponse := response.JSON200.Data[0]
	light := &smarthome.Light{
		ID:         lightResponse.Id,
		Name:       lightResponse.Metadata.Name,
		On:         lightResponse.On.On,
		Brightness: brightnessValue(lightResponse.Dimming.Brightness),
		Location:   s.host.ResolveLocationID(s.config.LocationID()),
	}
	return light, nil
}

func (s *huebridgeService) SetLight(ctx context.Context, id string, on *bool, brightness *uint8) (*smarthome.Light, error) {
	return nil, nil
}
