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
	"fmt"
	"log/slog"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tdrn-org/totem/internal/adapters/integration"
	"github.com/tdrn-org/totem/internal/domain/media"
)

func registerMediaTools(server *mcp.Server, runtime Runtime) {
	services := filterServices[media.MediaService](runtime.Services(), integration.NatureMedia, runtime.Logger())
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
		Description: "Search for media items (songs, albums, artists) across media servers. Specify genres like 'Rock', 'Pop', 'Alternative', or leave empty to search all.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query":  map[string]any{"type": "string", "description": "The search query."},
				"genres": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Filter by genres (e.g. ['Rock', 'Pop']). Empty = no filter."},
				"limit":  map[string]any{"type": "integer", "description": "Maximum number of results (default: 20)."},
			},
			"required": []string{"query"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		Query  string   `json:"query"`
		Genres []string `json:"genres,omitempty"`
		Limit  int      `json:"limit,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		if input.Limit <= 0 {
			input.Limit = 20
		}
		for _, ms := range services {
			items, err := ms.SearchItems(ctx, input.Query, input.Genres, input.Limit)
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
		line := "  " + item.Name
		if len(item.Artists) > 0 {
			// Artists is empty for MusicArtist/MusicAlbum items — printing the
			// slice directly yields a bare "[]".
			line += " — " + strings.Join(item.Artists, ", ")
		}
		// Duration.String() carries fractional seconds (3m46.1234567s); for
		// MusicArtist/MusicAlbum items it is the catalogue sum. Show m:ss.
		secs := int(item.Duration.Seconds())
		line += fmt.Sprintf(" (%d:%02d) [%s]\n", secs/60, secs%60, item.Type)
		result += line
		if len(item.Genres) > 0 {
			result += fmt.Sprintf("    Genres: %v\n", strings.Join(item.Genres, ", "))
		}
	}
	return result
}
