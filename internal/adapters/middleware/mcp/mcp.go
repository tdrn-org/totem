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
	"net/http"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tdrn-org/totem/internal/adapters/integration"
	"github.com/tdrn-org/totem/internal/buildinfo"
	"github.com/tdrn-org/totem/internal/domain/smarthome"
)

// Runtime provides the MCP handler access to the server's services and configuration.
type Runtime interface {
	BaseURL() *url.URL
	Logger() *slog.Logger
	Services() map[string]integration.Service
}

func NewHandler(runtime Runtime) http.Handler {
	impl := &mcp.Implementation{
		Name:       buildinfo.Cmd(),
		Version:    buildinfo.Version(),
		WebsiteURL: runtime.BaseURL().String(),
	}
	options := &mcp.ServerOptions{
		Logger: runtime.Logger(),
	}
	server := mcp.NewServer(impl, options)

	// Register tools from services based on their natures
	registerSmarthomeTools(server, runtime)

	getServer := func(_ *http.Request) *mcp.Server {
		return server
	}
	return mcp.NewStreamableHTTPHandler(getServer, nil)
}

func registerSmarthomeTools(server *mcp.Server, runtime Runtime) {
	var services []smarthome.SmarthomeService
	for _, service := range runtime.Services() {
		for _, nature := range service.Natures() {
			if nature != integration.NatureSmarthome {
				continue
			}
			shs, ok := service.(smarthome.SmarthomeService)
			if !ok {
				runtime.Logger().Warn("service declares smarthome nature but does not implement SmarthomeService",
					slog.String("serviceID", service.ID()))
				continue
			}
			services = append(services, shs)
		}
	}
	if len(services) == 0 {
		return
	}
	registerSmarthomeAPI(server, services, runtime.Logger())
}

func registerSmarthomeAPI(server *mcp.Server, services []smarthome.SmarthomeService, logger *slog.Logger) {
	// listLights: aggregate across all services
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listLights",
		Description: "List all lights known to the smart home system. Returns each light's ID, name, on/off state, brightness (0-100), and location.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		var allLights []*smarthome.Light
		for _, shs := range services {
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
			text := fmt.Sprintf("Light '%s' (%s): %s, brightness %d%%",
				light.Name, light.ID, onOffText(light.On), light.Brightness)
			if light.Location != nil {
				text += fmt.Sprintf("\nLocation: %s (%s)", light.Location.Name, light.Location.Address)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"light": light}, nil
		}
		return nil, nil, fmt.Errorf("light '%s' not found on any service", input.ID)
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
			text := fmt.Sprintf("Light '%s': %s, brightness %d%%",
				light.Name, onOffText(light.On), light.Brightness)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"light": light}, nil
		}
		return nil, nil, fmt.Errorf("light '%s' not found on any service", input.ID)
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
