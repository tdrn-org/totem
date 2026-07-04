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

package presence

import (
	"context"
	"time"

	"github.com/tdrn-org/totem/internal/domain"
	"github.com/tdrn-org/totem/internal/domain/smarthome"
)

// UserPresence describes the current and last known presence of a user.
type UserPresence struct {
	// Location is the user's current location, or nil if not currently present.
	Location *domain.Location
	// LastSeenLocation is the last known location of the user, regardless of current presence.
	LastSeenLocation *domain.Location
	// LastSeen is the timestamp of the most recent presence log entry, or nil if never seen.
	LastSeen *time.Time
}

// PresenceResult describes the current presence situation at a location.
type PresenceResult struct {
	// Users lists all users currently present at the location.
	Users []*domain.User
	// MotionSensors lists all motion sensors that are currently triggered.
	MotionSensors []*smarthome.MotionSensor
}

// PresenceService determines user presence across locations by correlating
// LAN device activity (MAC addresses) with registered user devices and
// motion sensor state.
type PresenceService interface {
	// CheckUser determines the current and last known location of the given user
	// by checking whether any of the user's registered devices are active on a
	// known LAN. Returns the full presence information including last seen data.
	CheckUser(ctx context.Context, user *domain.User) (*UserPresence, error)

	// CheckLocation returns all users currently present at the given location,
	// along with any motion sensors that are currently triggered.
	CheckLocation(ctx context.Context, location *domain.Location) (*PresenceResult, error)
}
