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
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/totem/internal/data/model"
)

func TestUser(t *testing.T) {
	driver := newTestDB(t)
	defer driver.Close()

	// Add user
	user1, err := model.AddUser(t.Context(), driver, "user1", "default")
	require.NoError(t, err)

	// Add user device
	device1MAC, _ := net.ParseMAC("01:02:03:04:05:06")
	device1, err := user1.AddDevice(t.Context(), device1MAC, "Device1")
	require.NoError(t, err)
	require.Equal(t, device1MAC.String(), device1.MAC)

	// Select all users (1)
	users, err := model.SelectUsers(t.Context(), driver)
	require.NoError(t, err)
	require.Len(t, users, 1)
	user2 := users[0]
	require.Equal(t, user1, user2)

	// Delete user
	err = user2.Delete(t.Context())
	require.NoError(t, err)

	// Select all users (0)
	users, err = model.SelectUsers(t.Context(), driver)
	require.NoError(t, err)
	require.Len(t, users, 0)
}
