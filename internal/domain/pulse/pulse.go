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
// Any number of pending pulses may exist per agent. A pulse is removed
// once it has been delivered, so the store holds pending pulses only.
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
	// SavePulse stores a new pending pulse for the given agent. Any
	// number of pending pulses may exist per agent.
	SavePulse(ctx context.Context, pulse *Pulse) error

	// ListAgentPulses returns all pending pulses for an agent, ordered
	// by scheduled time.
	ListAgentPulses(ctx context.Context, agent string) ([]*Pulse, error)

	// ListPulses returns all pulses, ordered by scheduled time.
	ListPulses(ctx context.Context) ([]*Pulse, error)

	// DeletePulse removes a pulse by ID. This is the standard way to
	// cancel a pulse; the ID is obtained from ListAgentPulses.
	DeletePulse(ctx context.Context, id string) error
}
