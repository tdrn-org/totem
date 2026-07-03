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

package data_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/go-database"
	"github.com/tdrn-org/go-database/memory"
	"github.com/tdrn-org/totem/internal/data"
	"github.com/tdrn-org/totem/internal/data/model"
	"github.com/tdrn-org/totem/internal/domain"
)

func TestUpdateScheme(t *testing.T) {
	store := newDataStore(t)
	err := store.Close()
	require.NoError(t, err)
}

func TestUser(t *testing.T) {
	store := newDataStore(t)
	defer store.Close()

	// Add user
	user1, err := store.AddUser(t.Context(), "user1", dummyLocation)
	require.NoError(t, err)

	// Add user device
	device1MAC, _ := net.ParseMAC("01:02:03:04:05:06")
	_, err = store.AddUserDevice(t.Context(), user1, device1MAC, "device1")
	require.NoError(t, err)

	// Update user
	err = store.UpdateUser(t.Context(), user1)
	require.NoError(t, err)

	// Delete user
	err = store.DeleteUser(t.Context(), user1)
	require.NoError(t, err)
}

func newDataStore(t *testing.T) *data.Store {
	driver, err := database.Open(memory.NewConfig(model.SqliteSchemaScriptOption))
	require.NoError(t, err)
	from, to, err := driver.UpdateSchema(t.Context())
	require.NoError(t, err)
	require.Equal(t, database.SchemaNone, from)
	require.Equal(t, 1, to)
	return data.NewStore(&dummyRuntime{}, driver, t.Name())
}

var dummyLocation *domain.Location = &domain.Location{
	ID:   "default",
	Name: "Default",
}

type dummyRuntime struct{}

func (runtime *dummyRuntime) ResolveLocationID(locationID string) *domain.Location {
	return dummyLocation
}
