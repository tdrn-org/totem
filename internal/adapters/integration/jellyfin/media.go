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

package jellyfin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	jellyfinapi "github.com/sj14/jellyfin-go/api"
	"github.com/tdrn-org/totem/internal/domain/media"
)

var includePlaylists []jellyfinapi.BaseItemKind = []jellyfinapi.BaseItemKind{jellyfinapi.BASEITEMKIND_PLAYLIST}

func (s jellyfinService) ListPlaylists(ctx context.Context) ([]*media.PlaylistSummary, error) {
	apiResponse, httpResponse, err := s.apiClient.ItemsAPI.
		GetItems(ctx).
		UserId(s.config.
			UserID()).
		IncludeItemTypes(includePlaylists).
		Recursive(true).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed request playlists from Jellyfin (cause: %w)", err)
	}
	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get playlists from Jellyfin (status: %s)", httpResponse.Status)
	}
	playlists := make([]*media.PlaylistSummary, 0, len(apiResponse.Items))
	for _, item := range apiResponse.Items {
		playlist := &media.PlaylistSummary{
			ID:        *item.Id,
			Name:      *item.Name.Get(),
			Duration:  time.Duration(item.GetRunTimeTicks()),
			ItemCount: int(item.GetChildCount()),
		}
		playlists = append(playlists, playlist)
	}
	return playlists, nil
}

func (s jellyfinService) GetPlaylist(ctx context.Context, id string) (*media.Playlist, error) {
	return nil, nil
}

func (s jellyfinService) AddToPlaylist(ctx context.Context, playlistID string, itemIDs []string) (*media.Playlist, error) {
	return nil, nil
}

func (s jellyfinService) RemoveFromPlaylist(ctx context.Context, playlistID string, itemIDs []string) (*media.Playlist, error) {
	return nil, nil
}

func (s jellyfinService) SearchItems(ctx context.Context, query string, mediaTypes []string, limit int) ([]*media.MediaItem, error) {
	return nil, nil
}

func (s jellyfinService) GetRecentlyPlayed(ctx context.Context, limit int) ([]*media.MediaItem, error) {
	return nil, nil
}
