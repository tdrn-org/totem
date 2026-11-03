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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/totem/internal/domain/pulse"
)

func testPulses() []*pulse.Pulse {
	now := time.Now()
	return []*pulse.Pulse{
		{ID: "p1", Agent: "judy", Message: "first", ScheduledAt: now.Add(time.Hour), CreatedAt: now},
		{ID: "p2", Agent: "judy", Message: "second", ScheduledAt: now.Add(2 * time.Hour), CreatedAt: now},
		{ID: "p3", Agent: "judy", Message: "third", ScheduledAt: now.Add(3 * time.Hour), CreatedAt: now},
	}
}

// TestPulsesToCancelByID verifies that a given id selects exactly that pulse.
func TestPulsesToCancelByID(t *testing.T) {
	selected, err := pulsesToCancel(testPulses(), "p2", false)
	require.NoError(t, err)
	require.Len(t, selected, 1)
	require.Equal(t, "p2", selected[0].ID)
}

// TestPulsesToCancelRefusesBulkWithoutConfirmation is the core of the fix: omitting
// both id and all must never delete anything — it must fail loudly instead.
func TestPulsesToCancelRefusesBulkWithoutConfirmation(t *testing.T) {
	selected, err := pulsesToCancel(testPulses(), "", false)
	require.Error(t, err)
	require.Nil(t, selected)
	// The message must name the affected pulses so the caller can decide.
	require.Contains(t, err.Error(), "3")
	require.Contains(t, err.Error(), "all=true")
}

// TestPulsesToCancelAllConfirmed verifies that an explicit all=true performs the
// bulk cancellation the caller asked for.
func TestPulsesToCancelAllConfirmed(t *testing.T) {
	selected, err := pulsesToCancel(testPulses(), "", true)
	require.NoError(t, err)
	require.Len(t, selected, 3)
}

// TestPulsesToCancelUnknownID verifies that a non-matching id is an error, not a
// silent no-op.
func TestPulsesToCancelUnknownID(t *testing.T) {
	selected, err := pulsesToCancel(testPulses(), "does-not-exist", false)
	require.Error(t, err)
	require.Nil(t, selected)
	require.Contains(t, err.Error(), "does-not-exist")
}

// TestPulsesToCancelIDAndAll verifies that the two mutually exclusive intents are
// rejected together instead of one silently winning.
func TestPulsesToCancelIDAndAll(t *testing.T) {
	selected, err := pulsesToCancel(testPulses(), "p1", true)
	require.Error(t, err)
	require.Nil(t, selected)
}

// TestPulsesToCancelNothingPending verifies that an empty pulse list is not an
// error — there is simply nothing to do.
func TestPulsesToCancelNothingPending(t *testing.T) {
	selected, err := pulsesToCancel(nil, "", false)
	require.NoError(t, err)
	require.Empty(t, selected)
}
