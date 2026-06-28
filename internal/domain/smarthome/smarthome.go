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

// Package smarthome defines the domain model for smart-home resources.
//
// The central abstraction is [SmarthomeService] — an interface that
// smart-home adapters (Hue, FRITZ!Box, etc.) implement. Consumers (MCP tools,
// REST handlers) work exclusively against this interface.
package smarthome

import (
	"context"

	"github.com/tdrn-org/totem/internal/domain"
)

// Light represents a single smart light.
type Light struct {
	// ID is the bridge-assigned unique identifier for this light.
	ID string
	// Name is the human-readable name of this light.
	Name string
	// On indicates whether the light is currently on.
	On bool
	// Brightness is the current brightness as a percentage (0-100).
	// 0 means the light is off or at minimum brightness.
	Brightness uint8
	// Location references the physical location this light belongs to.
	Location *domain.Location
}

// SmarthomeService defines the operations that any smart-home provider
// (Hue bridge, TR-064, etc.) must support for lighting control.
//
// Each registered integration service that implements this interface
// will have its lighting capabilities exposed as MCP tools.
type SmarthomeService interface {
	// ListLights returns all lights known to this service.
	ListLights(ctx context.Context) ([]*Light, error)

	// GetLight returns a single light by its bridge-assigned ID.
	// Returns an error if the light is not found.
	GetLight(ctx context.Context, id string) (*Light, error)

	// SetLight updates the state of a single light.
	// Fields set to nil are left unchanged; non-nil fields are applied.
	// Returns the updated light state.
	SetLight(ctx context.Context, id string, on *bool, brightness *uint8) (*Light, error)
}
