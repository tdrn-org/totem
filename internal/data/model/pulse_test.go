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

package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/totem/internal/data/model"
	"github.com/tdrn-org/totem/internal/domain/pulse"
)

func TestPulse(t *testing.T) {
	driver := newTestDB(t)
	defer driver.Close()

	pulse := &pulse.Pulse{
		Agent:       "agent",
		Message:     "Hello, ping",
		ScheduledAt: time.Now().Add(time.Hour),
	}

	// Insert pulse
	pulse1, err := model.InsertPulse(t.Context(), driver, pulse)
	require.NoError(t, err)
	require.NotNil(t, pulse1)

	// Select pulses by agent
	pulses, err := model.SelectPulsesByAgent(t.Context(), driver, pulse.Agent)
	require.NoError(t, err)
	require.Len(t, pulses, 1)
	require.Equal(t, pulse1, pulses[0])

	// Select pulses
	pulses, err = model.SelectPulsesByAgent(t.Context(), driver, pulse.Agent)
	require.NoError(t, err)
	require.Len(t, pulses, 1)
	require.Equal(t, pulse1, pulses[0])

	// Delete by id
	err = model.DeletePulsesByID(t.Context(), driver, pulse.ID)
	require.NoError(t, err)

	// Select pulses
	pulses, err = model.SelectPulsesByAgent(t.Context(), driver, pulse.Agent)
	require.NoError(t, err)
	require.Len(t, pulses, 0)

	// Insert pulse
	pulse1, err = model.InsertPulse(t.Context(), driver, pulse)
	require.NoError(t, err)
	require.NotNil(t, pulse1)

	// Delete by agent
	err = model.DeletePulsesByAgent(t.Context(), driver, pulse.Agent)
	require.NoError(t, err)

	// Select pulses
	pulses, err = model.SelectPulsesByAgent(t.Context(), driver, pulse.Agent)
	require.NoError(t, err)
	require.Len(t, pulses, 0)
}
