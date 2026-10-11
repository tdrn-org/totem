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
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tdrn-org/totem/internal/adapters/integration"
	"github.com/tdrn-org/totem/internal/domain/smarthome"
)

func registerSmarthomeTools(server *mcp.Server, runtime Runtime) {
	services := filterServices[smarthome.SmarthomeService](runtime.Services(), integration.NatureSmarthome, runtime.Logger())
	if len(services) == 0 {
		return
	}
	registerSmarthomeAPI(server, services, runtime)
}

func filterServicesByLocation(services []smarthome.SmarthomeService, locationID string, runtime Runtime) []smarthome.SmarthomeService {
	if locationID == "" {
		return services
	}
	location := runtime.ResolveLocationID(locationID)
	filtered := make([]smarthome.SmarthomeService, 0, len(services))
	for _, shs := range services {
		if shs.Location().Equal(location) {
			filtered = append(filtered, shs)
		}
	}
	return filtered
}

func registerSmarthomeAPI(server *mcp.Server, services []smarthome.SmarthomeService, runtime Runtime) {
	logger := runtime.Logger()
	// listLights: aggregate across all services, optionally filtered by location
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listLights",
		Description: "List all lights known to the smart home system. Returns each light's ID, name, on/off state, brightness (0-100), and location. Optionally filter by locationID.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"locationID": map[string]any{"type": "string", "description": "Optional location ID to filter by (e.g. 'Giselher16'). Omit to list all locations."},
			},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		LocationID string `json:"locationID,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		filtered := filterServicesByLocation(services, input.LocationID, runtime)
		var allLights []*smarthome.Light
		for _, shs := range filtered {
			lights, err := shs.ListLights(ctx)
			if err != nil {
				logger.Warn("failed to list lights from service", slog.Any("err", err))
				continue
			}
			allLights = append(allLights, lights...)
		}
		text := formatLightsText(allLights)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, map[string]any{"lights": allLights}, nil
	})

	// getLight: try each service until found
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getLight",
		Description: "Get details for a single light by its ID.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "description": "The light's unique identifier."},
			},
			"required": []string{"id"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		ID string `json:"id"`
	}) (*mcp.CallToolResult, any, error) {
		for _, shs := range services {
			light, err := shs.GetLight(ctx, input.ID)
			if err != nil {
				continue
			}
			text := fmt.Sprintf("Light %q (%s): %s, brightness %d%%",
				light.Name, light.ID, onOffText(light.On), light.Brightness)
			if light.Location != nil {
				text += fmt.Sprintf("\nLocation: %s (%s)", light.Location.Name, light.Location.Address)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"light": light}, nil
		}
		return nil, nil, fmt.Errorf("light %q not found on any service", input.ID)
	})

	// setLight: try each service until found
	mcp.AddTool(server, &mcp.Tool{
		Name:        "setLight",
		Description: "Set the on/off state and/or brightness of a single light. Pass only the fields you want to change — nil fields are left unchanged. Brightness is 0-100.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":         map[string]any{"type": "string", "description": "The light's unique identifier."},
				"on":         map[string]any{"type": "boolean", "description": "Turn the light on (true) or off (false). Omit to leave unchanged."},
				"brightness": map[string]any{"type": "integer", "description": "Set brightness (0-100). Omit to leave unchanged."},
			},
			"required": []string{"id"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		ID         string `json:"id"`
		On         *bool  `json:"on,omitempty"`
		Brightness *uint8 `json:"brightness,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		for _, shs := range services {
			light, err := shs.SetLight(ctx, input.ID, input.On, input.Brightness)
			if err != nil {
				continue
			}
			text := fmt.Sprintf("Light %q: %s, brightness %d%%",
				light.Name, onOffText(light.On), light.Brightness)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"light": light}, nil
		}
		return nil, nil, fmt.Errorf("light %q not found on any service", input.ID)
	})

	// listTemperatureSensors: aggregate across all services, optionally filtered by location
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listTemperatureSensors",
		Description: "List all temperature sensors known to the smart home system. Returns each sensor's ID, name, type, temperature, and location. Optionally filter by locationID.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"locationID": map[string]any{"type": "string", "description": "Optional location ID to filter by (e.g. 'Giselher16'). Omit to list all locations."},
			},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		LocationID string `json:"locationID,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		filtered := filterServicesByLocation(services, input.LocationID, runtime)
		var allSensors []*smarthome.TemperatureSensor
		for _, shs := range filtered {
			sensors, err := shs.ListTemperatureSensors(ctx)
			if err != nil {
				logger.Warn("failed to list temperature sensors from service", slog.Any("err", err))
				continue
			}
			allSensors = append(allSensors, sensors...)
		}
		text := formatTemperatureSensorsText(allSensors)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, map[string]any{"temperatureSensors": allSensors}, nil
	})

	// getTemperatureSensor: try each service until found
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getTemperatureSensor",
		Description: "Get details for a single temperature sensor by its ID.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "description": "The temperature sensor's unique identifier."},
			},
			"required": []string{"id"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		ID string `json:"id"`
	}) (*mcp.CallToolResult, any, error) {
		for _, shs := range services {
			sensor, err := shs.GetTemperatureSensor(ctx, input.ID)
			if err != nil {
				continue
			}
			text := fmt.Sprintf("Temperature sensor %q (%s): type=%s, %.1f°C",
				sensor.Name, sensor.ID, sensor.Type, sensor.Temperature)
			if sensor.Location != nil {
				text += fmt.Sprintf("\nLocation: %s (%s)", sensor.Location.Name, sensor.Location.Address)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"temperatureSensor": sensor}, nil
		}
		return nil, nil, fmt.Errorf("temperature sensor %q not found on any service", input.ID)
	})

	// listBatteryDevices: aggregate across all services, optionally filtered by location
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listBatteryDevices",
		Description: "List all battery-powered devices known to the smart home system. Returns each device's ID, name, type, battery level (0-100), and location. Optionally filter by locationID.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"locationID": map[string]any{"type": "string", "description": "Optional location ID to filter by (e.g. 'Giselher16'). Omit to list all locations."},
			},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		LocationID string `json:"locationID,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		filtered := filterServicesByLocation(services, input.LocationID, runtime)
		var allDevices []*smarthome.BatteryDevice
		for _, shs := range filtered {
			devices, err := shs.ListBatteryDevices(ctx)
			if err != nil {
				logger.Warn("failed to list battery devices from service", slog.Any("err", err))
				continue
			}
			allDevices = append(allDevices, devices...)
		}
		text := formatBatteryDevicesText(allDevices)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, map[string]any{"batteryDevices": allDevices}, nil
	})

	// getBatteryDevice: try each service until found
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getBatteryDevice",
		Description: "Get details for a single battery-powered device by its ID.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "description": "The battery device's unique identifier."},
			},
			"required": []string{"id"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		ID string `json:"id"`
	}) (*mcp.CallToolResult, any, error) {
		for _, shs := range services {
			device, err := shs.GetBatteryDevice(ctx, input.ID)
			if err != nil {
				continue
			}
			text := fmt.Sprintf("Battery device %q (%s): type=%s, 🔋%d%%",
				device.Name, device.ID, device.Type, device.BatteryLevel)
			if device.Location != nil {
				text += fmt.Sprintf("\nLocation: %s (%s)", device.Location.Name, device.Location.Address)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"batteryDevice": device}, nil
		}
		return nil, nil, fmt.Errorf("battery device %q not found on any service", input.ID)
	})

	// listMotionSensors: aggregate across all services, optionally filtered by location
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listMotionSensors",
		Description: "List all motion sensors known to the smart home system. Returns each sensor's ID, name, motion state (detected/clear), and location. Optionally filter by locationID.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"locationID": map[string]any{"type": "string", "description": "Optional location ID to filter by (e.g. 'Giselher16'). Omit to list all locations."},
			},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		LocationID string `json:"locationID,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		filtered := filterServicesByLocation(services, input.LocationID, runtime)
		var allMotionSensors []*smarthome.MotionSensor
		for _, shs := range filtered {
			sensors, err := shs.ListMotionSensors(ctx)
			if err != nil {
				logger.Warn("failed to list motion sensors from service", slog.Any("err", err))
				continue
			}
			allMotionSensors = append(allMotionSensors, sensors...)
		}
		text := formatMotionSensorsText(allMotionSensors)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, map[string]any{"motionSensors": allMotionSensors}, nil
	})

	// getMotion: try each service until found
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getMotion",
		Description: "Get motion details for a single motion sensor by its ID. Returns the sensor's name, motion state (detected/clear), and location.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "description": "The motion sensor's unique identifier."},
			},
			"required": []string{"id"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		ID string `json:"id"`
	}) (*mcp.CallToolResult, any, error) {
		for _, shs := range services {
			sensor, err := shs.GetMotion(ctx, input.ID)
			if err != nil {
				continue
			}
			motionText := "clear"
			if sensor.Motion {
				motionText = "detected 🏃"
			}
			text := fmt.Sprintf("Motion sensor %q (%s): motion=%s",
				sensor.Name, sensor.ID, motionText)
			if sensor.Location != nil {
				text += fmt.Sprintf("\nLocation: %s (%s)", sensor.Location.Name, sensor.Location.Address)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"motionSensor": sensor}, nil
		}
		return nil, nil, fmt.Errorf("motion sensor %q not found on any service", input.ID)
	})
}

func formatLightsText(lights []*smarthome.Light) string {
	if len(lights) == 0 {
		return "No lights found."
	}
	type summary struct {
		on  int
		off int
	}
	byLocation := make(map[string]*summary)
	for _, l := range lights {
		locName := "Unknown"
		if l.Location != nil && l.Location.Name != "" {
			locName = l.Location.Name
		}
		s, ok := byLocation[locName]
		if !ok {
			s = &summary{}
			byLocation[locName] = s
		}
		if l.On {
			s.on++
		} else {
			s.off++
		}
	}
	result := fmt.Sprintf("%d lights total:\n", len(lights))
	for loc, s := range byLocation {
		result += fmt.Sprintf("  %s: %d on, %d off\n", loc, s.on, s.off)
	}
	details, _ := json.MarshalIndent(lights, "", "  ")
	result += "\nDetails:\n" + string(details)
	return result
}

func onOffText(on bool) string {
	if on {
		return "ON"
	}
	return "OFF"
}

func formatTemperatureSensorsText(sensors []*smarthome.TemperatureSensor) string {
	if len(sensors) == 0 {
		return "No temperature sensors found."
	}
	type locSummary struct {
		count int
	}
	byLocation := make(map[string]*locSummary)
	for _, s := range sensors {
		locName := "Unknown"
		if s.Location != nil && s.Location.Name != "" {
			locName = s.Location.Name
		}
		l, ok := byLocation[locName]
		if !ok {
			l = &locSummary{}
			byLocation[locName] = l
		}
		l.count++
	}
	result := fmt.Sprintf("%d temperature sensors total:\n", len(sensors))
	for loc, l := range byLocation {
		result += fmt.Sprintf("  %s: %d sensors\n", loc, l.count)
	}
	result += "\n"
	for _, s := range sensors {
		result += fmt.Sprintf("  %-30s %-12s %5.1f°C\n", s.Name, s.Type, s.Temperature)
	}
	return result
}

func formatBatteryDevicesText(devices []*smarthome.BatteryDevice) string {
	if len(devices) == 0 {
		return "No battery devices found."
	}
	type locSummary struct {
		count int
	}
	byLocation := make(map[string]*locSummary)
	for _, d := range devices {
		locName := "Unknown"
		if d.Location != nil && d.Location.Name != "" {
			locName = d.Location.Name
		}
		l, ok := byLocation[locName]
		if !ok {
			l = &locSummary{}
			byLocation[locName] = l
		}
		l.count++
	}
	result := fmt.Sprintf("%d battery devices total:\n", len(devices))
	for loc, l := range byLocation {
		result += fmt.Sprintf("  %s: %d devices\n", loc, l.count)
	}
	result += "\n"
	for _, d := range devices {
		result += fmt.Sprintf("  %-30s %-12s 🔋%d%%\n", d.Name, d.Type, d.BatteryLevel)
	}
	return result
}

func formatMotionSensorsText(sensors []*smarthome.MotionSensor) string {
	if len(sensors) == 0 {
		return "No motion sensors found."
	}
	result := fmt.Sprintf("%d motion sensors:\n", len(sensors))
	for _, s := range sensors {
		motionText := "clear"
		if s.Motion {
			motionText = "detected 🏃"
		}
		locName := "Unknown"
		if s.Location != nil && s.Location.Name != "" {
			locName = s.Location.Name
		}
		result += fmt.Sprintf("  %-30s %-10s %s\n", s.Name, motionText, locName)
	}
	return result
}
