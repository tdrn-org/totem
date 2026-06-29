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

	jellyfinapi "github.com/sj14/jellyfin-go/api"
	"github.com/tdrn-org/totem/internal/domain/media"
)

var includePlaylists []jellyfinapi.BaseItemKind = []jellyfinapi.BaseItemKind{jellyfinapi.BASEITEMKIND_PLAYLIST}
var includeAudio []jellyfinapi.BaseItemKind = []jellyfinapi.BaseItemKind{jellyfinapi.BASEITEMKIND_AUDIO}

var filterIsPlayed []jellyfinapi.ItemFilter = []jellyfinapi.ItemFilter{jellyfinapi.ITEMFILTER_IS_PLAYED}

func (s *jellyfinService) ListPlaylists(ctx context.Context) ([]*media.PlaylistSummary, error) {
	apiResponse, httpResponse, err := s.apiClient.ItemsAPI.
		GetItems(ctx).
		UserId(s.config.UserID()).
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
			Duration:  ticksToDuration(item.GetRunTimeTicks()),
			ItemCount: int(item.GetChildCount()),
		}
		playlists = append(playlists, playlist)
	}
	return playlists, nil
}

func (s *jellyfinService) GetPlaylist(ctx context.Context, id string) (*media.Playlist, error) {
	apiResponse, httpResponse, err := s.apiClient.UserLibraryAPI.
		GetItem(ctx, id).
		UserId(s.config.UserID()).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed request playlist from Jellyfin (cause: %w)", err)
	}
	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get playlist from Jellyfin (status: %s)", httpResponse.Status)
	}
	mediaItems, err := s.getPlaylistItems(ctx, id)
	if err != nil {
		return nil, err
	}
	playlist := &media.Playlist{
		ID:        *apiResponse.Id,
		Name:      apiResponse.GetName(),
		Duration:  ticksToDuration(apiResponse.GetRunTimeTicks()),
		ItemCount: len(mediaItems),
		Items:     mediaItems,
	}
	return playlist, nil
}

func (s *jellyfinService) getPlaylistItems(ctx context.Context, id string) ([]*media.MediaItem, error) {
	apiResponse, httpResponse, err := s.apiClient.PlaylistsAPI.
		GetPlaylistItems(ctx, id).
		UserId(s.config.UserID()).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed request playlist items from Jellyfin (cause: %w)", err)
	}
	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get playlist items from Jellyfin (status: %s)", httpResponse.Status)
	}
	mediaItems := make([]*media.MediaItem, 0, len(apiResponse.Items))
	for _, item := range apiResponse.Items {
		mediaItem := &media.MediaItem{
			ID:       *item.Id,
			Name:     *item.Name.Get(),
			Type:     string(*item.Type),
			Artists:  item.Artists,
			Album:    *item.Album.Get(),
			Duration: ticksToDuration(item.GetRunTimeTicks()),
			Genres:   item.Genres,
			// TODO: ImageURL: "",
		}
		mediaItems = append(mediaItems, mediaItem)
	}
	return mediaItems, nil
}

func (s *jellyfinService) AddToPlaylist(ctx context.Context, playlistID string, itemIDs []string) (*media.Playlist, error) {
	httpResponse, err := s.apiClient.PlaylistsAPI.
		AddItemToPlaylist(ctx, playlistID).
		UserId(s.config.UserID()).
		Ids(itemIDs).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to request add to playlist from Jellyfin (cause: %w)", err)
	}
	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to add playlist items from Jellyfin (status: %s)", httpResponse.Status)
	}
	return s.GetPlaylist(ctx, playlistID)
}

func (s *jellyfinService) RemoveFromPlaylist(ctx context.Context, playlistID string, itemIDs []string) (*media.Playlist, error) {
	httpResponse, err := s.apiClient.PlaylistsAPI.
		RemoveItemFromPlaylist(ctx, playlistID).
		EntryIds(itemIDs).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to request add to playlist from Jellyfin (cause: %w)", err)
	}
	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to add playlist items from Jellyfin (status: %s)", httpResponse.Status)
	}
	return s.GetPlaylist(ctx, playlistID)
}

func (s *jellyfinService) SearchItems(ctx context.Context, query string, mediaTypes []string, limit int) ([]*media.MediaItem, error) {
	// TODO: Use mediaTypes for filtering
	apiResponse, httpResponse, err := s.apiClient.ItemsAPI.
		GetItems(ctx).
		UserId(s.config.UserID()).
		IncludeItemTypes(includeAudio).
		SearchTerm(query).
		Recursive(true).
		Limit(int32(limit)).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed request items from Jellyfin (cause: %w)", err)
	}
	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get items from Jellyfin (status: %s)", httpResponse.Status)
	}
	mediaItems := make([]*media.MediaItem, 0, len(apiResponse.Items))
	for _, item := range apiResponse.Items {
		mediaItem := &media.MediaItem{
			ID:       *item.Id,
			Name:     item.GetName(),
			Type:     string(*item.Type),
			Artists:  item.Artists,
			Album:    *item.Album.Get(),
			Duration: ticksToDuration(item.GetRunTimeTicks()),
			Genres:   item.Genres,
			// TODO: ImageURL: "",
		}
		mediaItems = append(mediaItems, mediaItem)
	}
	return mediaItems, nil
}

func (s *jellyfinService) GetRecentlyPlayed(ctx context.Context, limit int) ([]*media.MediaItem, error) {
	apiResponse, httpResponse, err := s.apiClient.ItemsAPI.
		GetItems(ctx).
		UserId(s.config.UserID()).
		IncludeItemTypes(includeAudio).
		Filters(filterIsPlayed).
		Recursive(true).
		SortBy([]jellyfinapi.ItemSortBy{jellyfinapi.ITEMSORTBY_DATE_PLAYED}).
		SortOrder([]jellyfinapi.SortOrder{jellyfinapi.SORTORDER_DESCENDING}).
		Limit(int32(limit)).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed request recently played from Jellyfin (cause: %w)", err)
	}
	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get recently played from Jellyfin (status: %s)", httpResponse.Status)
	}
	mediaItems := make([]*media.MediaItem, 0, len(apiResponse.Items))
	for _, item := range apiResponse.Items {
		mediaItem := &media.MediaItem{
			ID:       *item.Id,
			Name:     item.GetName(),
			Type:     string(*item.Type),
			Artists:  item.Artists,
			Album:    *item.Album.Get(),
			Duration: ticksToDuration(item.GetRunTimeTicks()),
			Genres:   item.Genres,
			// TODO: ImageURL: "",
		}
		mediaItems = append(mediaItems, mediaItem)
	}
	return mediaItems, nil
}
