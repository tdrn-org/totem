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

	"github.com/tdrn-org/totem/internal/domain"
	"github.com/tdrn-org/totem/internal/domain/smarthome"
)

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
	// CheckUser determines the current location of the given user by checking
	// whether any of the user's registered devices are active on a known LAN.
	// Returns nil, nil if the user cannot be located at any known location.
	CheckUser(ctx context.Context, user *domain.User) (*domain.Location, error)

	// CheckLocation returns all users currently present at the given location,
	// along with any motion sensors that are currently triggered.
	CheckLocation(ctx context.Context, location *domain.Location) (*PresenceResult, error)
}
