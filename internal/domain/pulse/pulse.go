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

// Package pulse defines the domain types and interfaces for Judy's
// echo_pulse feature — self-addressed delayed notifications.
//
// At most one pending pulse exists per agent at any time. Creating a
// new pulse overwrites any existing pending pulse for that agent.
package pulse

import (
	"context"
	"time"
)

// Pulse is a self-addressed delayed notification. An agent (e.g., Judy)
// creates a pulse to receive a reminder or check-in at a future time.
type Pulse struct {
	ID          string    `json:"id"`
	Agent       string    `json:"agent"`
	Message     string    `json:"message"`
	ScheduledAt time.Time `json:"scheduled_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// Store defines the persistence interface for echo pulses. Fired pulses
// are deleted after delivery — no status, no housekeeping.
type Store interface {
	// SavePulse stores a new pending pulse for the given agent. If a
	// pending pulse already exists for the agent, it is silently overwritten.
	SavePulse(ctx context.Context, pulse *Pulse) error

	// ListAgentPulses returns all saved pulses for an agent (pending + history).
	ListAgentPulses(ctx context.Context, agent string) ([]*Pulse, error)

	// ListPulses returns all pulses.
	ListPulses(ctx context.Context) ([]*Pulse, error)

	// DeletePulse removes a pulse by ID.
	DeletePulse(ctx context.Context, id string) error
}
