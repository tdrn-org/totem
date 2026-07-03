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
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/totem/internal/data/model"
)

func TestPresenceLog(t *testing.T) {
	driver := newTestDB(t)
	defer driver.Close()

	// Add logs
	log1, err := model.LogPresence(t.Context(), driver, "user1", "default", true, netip.Addr{}, "")
	require.NoError(t, err)
	log2, err := model.LogPresence(t.Context(), driver, "user1", "default", false, netip.Addr{}, "")
	require.NoError(t, err)
	log3, err := model.LogPresence(t.Context(), driver, "user1", "default", true, netip.Addr{}, "")
	require.NoError(t, err)

	// Select logs
	logs, err := model.SelectPresenceLogByUserID(t.Context(), driver, log1.UserID)
	require.NoError(t, err)
	require.Len(t, logs, 2)
	require.Equal(t, log2, logs[1])
	require.Equal(t, log3, logs[0])
}
