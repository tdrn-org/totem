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

package media

import (
	"context"
	"time"
)

// MediaService defines the operations for media servers.
// A media server provides access to music, playlists, and playback history.
// It is location-less — media follows the user wherever they are.
type MediaService interface {
	// ListPlaylists returns a summary of all playlists available on the media server.
	// Summaries omit items to avoid N+1 API calls — use GetPlaylist for details.
	ListPlaylists(ctx context.Context) ([]*PlaylistSummary, error)

	// GetPlaylist returns a single playlist including all its items.
	GetPlaylist(ctx context.Context, id string) (*Playlist, error)

	// CreatePlaylist creates an playlist with the given name and the given items.
	// At least one item must be added during create.
	CreatePlaylist(ctx context.Context, name string, itemIDs []string) (*Playlist, error)

	// AddToPlaylist adds one or more items to a playlist.
	// Returns the updated playlist.
	AddToPlaylist(ctx context.Context, playlistID string, itemIDs []string) (*Playlist, error)

	// RemoveFromPlaylist removes one or more items from a playlist.
	// Returns the updated playlist.
	RemoveFromPlaylist(ctx context.Context, playlistID string, itemIDs []string) (*Playlist, error)

	// SearchItems searches for media items matching the query.
	// genres can filter by genre (e.g. "Rock", "Pop", "Jazz").
	// An empty genres slice applies no genre filter.
	SearchItems(ctx context.Context, query string, genres []string, limit int) ([]*MediaItem, error)

	// GetRecentlyPlayed returns recently played items for the configured user.
	GetRecentlyPlayed(ctx context.Context, limit int) ([]*MediaItem, error)
}

// PlaylistSummary is a lightweight view of a playlist without items.
// Used by ListPlaylists to avoid N+1 API calls per playlist.
type PlaylistSummary struct {
	ID        string
	Name      string
	Duration  time.Duration
	ItemCount int
}

// Playlist represents a named, ordered collection of media items including the full item list.
// Use GetPlaylist to retrieve this detailed view.
type Playlist struct {
	ID        string
	Name      string
	Duration  time.Duration
	ItemCount int
	Items     []*MediaItem
}

// MediaItem represents a single media entity.
// Depending on Type, this can be a song ("Audio"), an album ("MusicAlbum"),
// or an artist ("MusicArtist").
type MediaItem struct {
	ID       string
	Name     string
	Type     string
	Artists  []string
	Album    string
	Duration time.Duration
	Genres   []string
	ImageURL string
}
