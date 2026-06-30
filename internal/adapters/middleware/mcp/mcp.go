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
	"github.com/tdrn-org/totem/internal/domain/media"
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
	registerMediaTools(server, runtime)

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

	// listSensors: aggregate across all services
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listSensors",
		Description: "List all sensors known to the smart home system. Returns each sensor's ID, name, type, temperature (if applicable), and location.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		var allSensors []*smarthome.Sensor
		for _, shs := range services {
			sensors, err := shs.ListSensors(ctx)
			if err != nil {
				logger.Warn("failed to list sensors from service", slog.Any("err", err))
				continue
			}
			allSensors = append(allSensors, sensors...)
		}
		text := formatSensorsText(allSensors)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, map[string]any{"sensors": allSensors}, nil
	})

	// getSensor: try each service until found
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getSensor",
		Description: "Get details for a single sensor by its ID.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "description": "The sensor's unique identifier."},
			},
			"required": []string{"id"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		ID string `json:"id"`
	}) (*mcp.CallToolResult, any, error) {
		for _, shs := range services {
			sensor, err := shs.GetSensor(ctx, input.ID)
			if err != nil {
				continue
			}
			text := fmt.Sprintf("Sensor '%s' (%s): type=%s, %.1f°C",
				sensor.Name, sensor.ID, sensor.Type, sensor.Temperature)
			if sensor.Location != nil {
				text += fmt.Sprintf("\nLocation: %s (%s)", sensor.Location.Name, sensor.Location.Address)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"sensor": sensor}, nil
		}
		return nil, nil, fmt.Errorf("sensor '%s' not found on any service", input.ID)
	})
}

func registerMediaTools(server *mcp.Server, runtime Runtime) {
	var services []media.MediaService
	for _, service := range runtime.Services() {
		for _, nature := range service.Natures() {
			if nature != integration.NatureMedia {
				continue
			}
			ms, ok := service.(media.MediaService)
			if !ok {
				runtime.Logger().Warn("service declares media nature but does not implement MediaService",
					slog.String("serviceID", service.ID()))
				continue
			}
			services = append(services, ms)
		}
	}
	if len(services) == 0 {
		return
	}
	registerMediaAPI(server, services, runtime.Logger())
}

func registerMediaAPI(server *mcp.Server, services []media.MediaService, logger *slog.Logger) {
	// listPlaylists: aggregate summaries across all media services
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listPlaylists",
		Description: "List all playlists from media servers. Returns each playlist's ID, name, total duration (as a Go duration string), and item count. Media is location-less.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		var allPlaylists []*media.PlaylistSummary
		for _, ms := range services {
			playlists, err := ms.ListPlaylists(ctx)
			if err != nil {
				logger.Warn("failed to list playlists from service", slog.Any("err", err))
				continue
			}
			allPlaylists = append(allPlaylists, playlists...)
		}
		text := formatPlaylistsText(allPlaylists)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, map[string]any{"playlists": allPlaylists}, nil
	})

	// getPlaylist: try each service until found
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getPlaylist",
		Description: "Get details for a single playlist including all its items. Use the playlist ID from listPlaylists.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "description": "The playlist's unique identifier."},
			},
			"required": []string{"id"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		ID string `json:"id"`
	}) (*mcp.CallToolResult, any, error) {
		for _, ms := range services {
			playlist, err := ms.GetPlaylist(ctx, input.ID)
			if err != nil {
				continue
			}
			text := fmt.Sprintf("Playlist '%s': %d items, %s total",
				playlist.Name, playlist.ItemCount, playlist.Duration)
			if len(playlist.Items) > 0 {
				text += "\n\nTracks:"
				for i, item := range playlist.Items {
					text += fmt.Sprintf("\n  %d. %s — %s (%s)", i+1, item.Name, item.Artists, item.Duration)
				}
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"playlist": playlist}, nil
		}
		return nil, nil, fmt.Errorf("playlist '%s' not found on any service", input.ID)
	})

	// addToPlaylist: try each service until found
	mcp.AddTool(server, &mcp.Tool{
		Name:        "addToPlaylist",
		Description: "Add one or more media items to a playlist. Use item IDs from searchMedia results. Returns the updated playlist.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"playlistID": map[string]any{"type": "string", "description": "The playlist's unique identifier."},
				"itemIDs":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "One or more media item IDs to add."},
			},
			"required": []string{"playlistID", "itemIDs"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		PlaylistID string   `json:"playlistID"`
		ItemIDs    []string `json:"itemIDs"`
	}) (*mcp.CallToolResult, any, error) {
		for _, ms := range services {
			playlist, err := ms.AddToPlaylist(ctx, input.PlaylistID, input.ItemIDs)
			if err != nil {
				continue
			}
			text := fmt.Sprintf("Added %d item(s) to playlist '%s'.\nNew total: %d items, %s",
				len(input.ItemIDs), playlist.Name, playlist.ItemCount, playlist.Duration)
			if len(playlist.Items) > 0 {
				text += "\n\nTracks:"
				for i, item := range playlist.Items {
					text += fmt.Sprintf("\n  %d. %s — %s (%s)", i+1, item.Name, item.Artists, item.Duration)
				}
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"playlist": playlist}, nil
		}
		return nil, nil, fmt.Errorf("playlist '%s' not found on any service", input.PlaylistID)
	})

	// createPlaylist: try each service until created (requires at least one initial item)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createPlaylist",
		Description: "Create a new playlist with the given name and at least one initial media item. Use item IDs from searchMedia results. Returns the created playlist.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":    map[string]any{"type": "string", "description": "The name for the new playlist."},
				"itemIDs": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "One or more media item IDs to seed the playlist (at least one required)."},
			},
			"required": []string{"name", "itemIDs"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		Name    string   `json:"name"`
		ItemIDs []string `json:"itemIDs"`
	}) (*mcp.CallToolResult, any, error) {
		for _, ms := range services {
			playlist, err := ms.CreatePlaylist(ctx, input.Name, input.ItemIDs)
			if err != nil {
				continue
			}
			text := fmt.Sprintf("Created playlist '%s' with %d item(s).\nDuration: %s",
				playlist.Name, playlist.ItemCount, playlist.Duration)
			if len(playlist.Items) > 0 {
				text += "\n\nTracks:"
				for i, item := range playlist.Items {
					text += fmt.Sprintf("\n  %d. %s — %s (%s)", i+1, item.Name, item.Artists, item.Duration)
				}
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"playlist": playlist}, nil
		}
		return nil, nil, fmt.Errorf("failed to create playlist '%s' on any service", input.Name)
	})

	// removeFromPlaylist: try each service until found
	mcp.AddTool(server, &mcp.Tool{
		Name:        "removeFromPlaylist",
		Description: "Remove one or more media items from a playlist. Use item IDs from getPlaylist results. Returns the updated playlist.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"playlistID": map[string]any{"type": "string", "description": "The playlist's unique identifier."},
				"itemIDs":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "One or more media item IDs to remove."},
			},
			"required": []string{"playlistID", "itemIDs"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		PlaylistID string   `json:"playlistID"`
		ItemIDs    []string `json:"itemIDs"`
	}) (*mcp.CallToolResult, any, error) {
		for _, ms := range services {
			playlist, err := ms.RemoveFromPlaylist(ctx, input.PlaylistID, input.ItemIDs)
			if err != nil {
				continue
			}
			text := fmt.Sprintf("Removed %d item(s) from playlist '%s'.\nNew total: %d items, %s",
				len(input.ItemIDs), playlist.Name, playlist.ItemCount, playlist.Duration)
			if len(playlist.Items) > 0 {
				text += "\n\nTracks:"
				for i, item := range playlist.Items {
					text += fmt.Sprintf("\n  %d. %s — %s (%s)", i+1, item.Name, item.Artists, item.Duration)
				}
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"playlist": playlist}, nil
		}
		return nil, nil, fmt.Errorf("playlist '%s' not found on any service", input.PlaylistID)
	})

	// searchMedia: try each service until results found
	mcp.AddTool(server, &mcp.Tool{
		Name:        "searchMedia",
		Description: "Search for media items (songs, albums, artists) across media servers. Specify types like 'Audio', 'MusicAlbum', 'MusicArtist', or leave empty to search all.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query":      map[string]any{"type": "string", "description": "The search query."},
				"mediaTypes": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Filter by media types (e.g. ['Audio', 'MusicAlbum']). Empty = all."},
				"limit":      map[string]any{"type": "integer", "description": "Maximum number of results (default: 20)."},
			},
			"required": []string{"query"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		Query      string   `json:"query"`
		MediaTypes []string `json:"mediaTypes,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		if input.Limit <= 0 {
			input.Limit = 20
		}
		for _, ms := range services {
			items, err := ms.SearchItems(ctx, input.Query, input.MediaTypes, input.Limit)
			if err != nil {
				continue
			}
			text := formatMediaItemsText(items)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"items": items}, nil
		}
		return nil, nil, fmt.Errorf("no results found for '%s'", input.Query)
	})

	// getRecentlyPlayed: try each service until results found
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getRecentlyPlayed",
		Description: "Get recently played media items for the current user.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"limit": map[string]any{"type": "integer", "description": "Maximum number of items (default: 10)."},
			},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		Limit int `json:"limit,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		if input.Limit <= 0 {
			input.Limit = 10
		}
		for _, ms := range services {
			items, err := ms.GetRecentlyPlayed(ctx, input.Limit)
			if err != nil {
				continue
			}
			text := formatMediaItemsText(items)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, map[string]any{"items": items}, nil
		}
		return nil, nil, fmt.Errorf("no recently played items found")
	})
}

func formatPlaylistsText(playlists []*media.PlaylistSummary) string {
	if len(playlists) == 0 {
		return "No playlists found."
	}
	result := fmt.Sprintf("%d playlists:\n", len(playlists))
	for _, p := range playlists {
		result += fmt.Sprintf("  %s | %s | %d items | %s\n", p.Name, p.ID, p.ItemCount, p.Duration)
	}
	return result
}

func formatMediaItemsText(items []*media.MediaItem) string {
	if len(items) == 0 {
		return "No items found."
	}
	result := fmt.Sprintf("%d items:\n", len(items))
	for _, item := range items {
		result += fmt.Sprintf("  %s — %s (%s) [%s]\n", item.Name, item.Artists, item.Duration, item.Type)
		if len(item.Genres) > 0 {
			result += fmt.Sprintf("    Genres: %v\n", item.Genres)
		}
	}
	return result
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

func formatSensorsText(sensors []*smarthome.Sensor) string {
	if len(sensors) == 0 {
		return "No sensors found."
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
	result := fmt.Sprintf("%d sensors total:\n", len(sensors))
	for loc, l := range byLocation {
		result += fmt.Sprintf("  %s: %d sensors\n", loc, l.count)
	}
	result += "\n"
	for _, s := range sensors {
		battery := ""
		if s.BatteryLevel >= 0 {
			battery = fmt.Sprintf(" 🔋%d%%", s.BatteryLevel)
		}
		result += fmt.Sprintf("  %-30s %-12s %5.1f°C%s\n", s.Name, s.Type, s.Temperature, battery)
	}
	return result
}
