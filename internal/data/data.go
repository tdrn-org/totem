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

package data

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"slices"

	"github.com/tdrn-org/go-database"
	"github.com/tdrn-org/totem/internal/data/model"
	"github.com/tdrn-org/totem/internal/domain"
)

type Runtime interface {
	// ResolveLocationID resolved the given location id to the location. This
	// function never fails, but returns the default location in case th id
	// cannot be resolved.
	ResolveLocationID(locationID string) *domain.Location
}

type Store struct {
	runtime Runtime
	driver  *database.Driver
	logger  *slog.Logger
}

func NewStore(runtime Runtime, driver *database.Driver, name string) *Store {
	return &Store{
		runtime: runtime,
		driver:  driver,
		logger:  slog.With(slog.String("name", name)),
	}
}

func (s *Store) Ping(ctx context.Context) error {
	return s.driver.Ping(ctx)
}

func (s *Store) Close() error {
	return s.driver.Close()
}

func (s *Store) ListUsers(ctx context.Context) ([]*domain.User, error) {
	storeUsers, err := model.SelectUsers(ctx, s.driver)
	if err != nil {
		return nil, err
	}
	users := make([]*domain.User, 0, len(storeUsers))
	for _, storeUser := range storeUsers {
		user, err := s.model2User(storeUser)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *Store) AddUser(ctx context.Context, name string, defaultLocation *domain.Location) (*domain.User, error) {
	storeUser, err := model.AddUser(ctx, s.driver, name, defaultLocation.ID)
	if err != nil {
		return nil, err
	}
	user, err := s.model2User(storeUser)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) AddUserDevice(ctx context.Context, user *domain.User, mac net.HardwareAddr, name string) (*domain.UserDevice, error) {
	storeUser := model.ExistingUser(s.driver, user.ID, user.Name, user.DefaultLocation.ID)
	device, err := storeUser.AddDevice(ctx, mac, name)
	if err != nil {
		return nil, err
	}
	return s.model2UserDevice(device)
}

func (s *Store) RemoveUserDevice(ctx context.Context, user *domain.User, mac net.HardwareAddr) error {
	n := 0
	for _, device := range user.Devices {
		if !slices.Equal(device.MAC, mac) {
			user.Devices[n] = device
			n++
		}
	}
	user.Devices = user.Devices[:n]
	return s.UpdateUser(ctx, user)
}

func (s *Store) UpdateUser(ctx context.Context, user *domain.User) error {
	modelUser := s.user2Model(user)
	return modelUser.Update(ctx)
}

func (s *Store) DeleteUser(ctx context.Context, user *domain.User) error {
	modelUser := s.user2Model(user)
	return modelUser.Delete(ctx)
}

func (s *Store) model2User(storeUser *model.User) (*domain.User, error) {
	user := &domain.User{
		ID:              storeUser.ID,
		Name:            storeUser.Name,
		DefaultLocation: s.runtime.ResolveLocationID(storeUser.DefaultLocationID),
		Devices:         make([]*domain.UserDevice, 0, len(storeUser.Devices)),
	}
	for _, storeDevice := range storeUser.Devices {
		device, err := s.model2UserDevice(storeDevice)
		if err != nil {
			return nil, err
		}
		user.Devices = append(user.Devices, device)
	}
	return user, nil
}

func (s *Store) model2UserDevice(storeDevice *model.UserDevice) (*domain.UserDevice, error) {
	deviceMAC, err := net.ParseMAC(storeDevice.MAC)
	if err != nil {
		return nil, fmt.Errorf("invalid MAC '%s' (cause: %w)", storeDevice.MAC, err)
	}
	device := &domain.UserDevice{
		MAC:  deviceMAC,
		Name: storeDevice.Name,
	}
	return device, nil
}

func (s *Store) user2Model(user *domain.User) *model.User {
	storeUser := model.ExistingUser(s.driver, user.ID, user.Name, user.DefaultLocation.ID)
	for _, device := range user.Devices {
		storeUser.ExistingDevice(device.MAC, device.Name)
	}
	return storeUser
}
