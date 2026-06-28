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

	"github.com/tdrn-org/go-hue"
	"github.com/tdrn-org/go-hue/api"
	"github.com/tdrn-org/totem/internal/domain/smarthome"
)

func (s *huebridgeService) ListLights(ctx context.Context) ([]*smarthome.Light, error) {
	bridgeClient, err := s.resolveBridgeClient()
	if err != nil {
		return nil, err
	}
	response, err := bridgeClient.GetLights(ctx)
	if err != nil {
		s.taintBridgeClient()
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
		s.taintBridgeClient()
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
	bridgeClient, err := s.resolveBridgeClient()
	if err != nil {
		return nil, err
	}
	body := api.LightPut{}
	if on != nil {
		body.On = &api.On{On: *on}
	}
	if brightness != nil {
		bri := api.Brightness(*brightness) / 100.0
		if bri <= 0.0 {
			bri = 0.01
		}
		if bri > 1.0 {
			bri = 1.0
		}
		body.Dimming = &api.Dimming{Brightness: bri}
	}
	_, err = bridgeClient.UpdateLight(ctx, id, body)
	if err != nil {
		s.taintBridgeClient()
		return nil, fmt.Errorf("failed to update light on Hue Bridge (cause: %w)", err)
	}
	return s.GetLight(ctx, id)
}

func (s *huebridgeService) ListSensors(ctx context.Context) ([]*smarthome.Sensor, error) {
	bridgeClient, err := s.resolveBridgeClient()
	if err != nil {
		return nil, err
	}
	location := s.host.ResolveLocationID(s.config.LocationID())
	deviceNames := s.resolveDeviceNames(ctx, bridgeClient)

	response, err := bridgeClient.GetTemperatures(ctx)
	if err != nil {
		s.taintBridgeClient()
		return nil, fmt.Errorf("failed to get temperatures from Hue Bridge (cause: %w)", err)
	}
	sensors := make([]*smarthome.Sensor, 0, len(response.JSON200.Data))
	for _, t := range response.JSON200.Data {
		temp := float64(t.Temperature.Temperature)
		if t.Temperature.TemperatureReport != nil && t.Temperature.TemperatureReport.Temperature != nil {
			temp = float64(*t.Temperature.TemperatureReport.Temperature)
		}
		sensor := &smarthome.Sensor{
			ID:          t.Id,
			Name:        deviceName(deviceNames, t.Owner.Rid, t.Id),
			Type:        "temperature",
			Temperature: temp,
			Location:    location,
		}
		sensors = append(sensors, sensor)
	}
	return sensors, nil
}

func (s *huebridgeService) GetSensor(ctx context.Context, id string) (*smarthome.Sensor, error) {
	bridgeClient, err := s.resolveBridgeClient()
	if err != nil {
		return nil, err
	}
	location := s.host.ResolveLocationID(s.config.LocationID())
	response, err := bridgeClient.GetTemperature(ctx, id)
	if err != nil {
		s.taintBridgeClient()
		return nil, fmt.Errorf("failed to get temperature from Hue Bridge (cause: %w)", err)
	}
	if len(response.JSON200.Data) == 0 {
		return nil, fmt.Errorf("no temperature data from Hue Bridge")
	}
	deviceNames := s.resolveDeviceNames(ctx, bridgeClient)
	t := response.JSON200.Data[0]
	temp := float64(t.Temperature.Temperature)
	if t.Temperature.TemperatureReport != nil && t.Temperature.TemperatureReport.Temperature != nil {
		temp = float64(*t.Temperature.TemperatureReport.Temperature)
	}
	return &smarthome.Sensor{
		ID:          t.Id,
		Name:        deviceName(deviceNames, t.Owner.Rid, t.Id),
		Type:        "temperature",
		Temperature: temp,
		Location:    location,
	}, nil
}

func (s *huebridgeService) resolveDeviceNames(ctx context.Context, client hue.BridgeClient) map[string]string {
	names := make(map[string]string)
	response, err := client.GetDevices(ctx)
	if err != nil {
		return names
	}
	for _, d := range response.JSON200.Data {
		names[d.Id] = d.Metadata.Name
	}
	return names
}

func deviceName(names map[string]string, ownerRID, fallback string) string {
	if name, ok := names[ownerRID]; ok && name != "" {
		return name
	}
	return fallback
}
