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

package model

import (
	"context"
	_ "embed"
	"net"

	"github.com/tdrn-org/go-database"
)

type User struct {
	ID                string `db:"id"`
	Name              string `db:"name"`
	DefaultLocationID string `db:"default_location_id"`
	Devices           []*UserDevice
}

type UserDevice struct {
	MAC    string `db:"mac"`
	Name   string `db:"name"`
	UserID string `db:"user_id"`
}

//go:embed user.insert.sql
var insertUserSQL string

// InsertUser inserts a new user without any devices.
func InsertUser(ctx context.Context, driver *database.Driver, name, defaultLocationID string) (*User, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	user := &User{
		ID:                database.NewID(),
		Name:              name,
		DefaultLocationID: defaultLocationID,
		Devices:           []*UserDevice{},
	}
	err = tx.ExecTx(txCtx, insertUserSQL, user.ID, user.Name, user.DefaultLocationID)
	if err != nil {
		return nil, err
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ExistingUser creates a new user without any devices, from already stored data
func ExistingUser(driver *database.Driver, id, name, defaultLocationID string) *User {
	return &User{
		ID:                id,
		Name:              name,
		DefaultLocationID: defaultLocationID,
		Devices:           []*UserDevice{},
	}
}

// ExistingDevice creates a new device for the given user, from already stored data
func (u *User) ExistingDevice(mac net.HardwareAddr, name string) *UserDevice {
	device := &UserDevice{
		MAC:    mac.String(),
		Name:   name,
		UserID: u.ID,
	}
	u.Devices = append(u.Devices, device)
	return device
}

//go:embed user.update_by_id.sql
var updateUserByIDSQL string

// Update updates the stored data for the user as well as any associated device
func (u *User) Update(ctx context.Context, driver *database.Driver) error {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	err = tx.ExecTx(txCtx, updateUserByIDSQL, u.Name, u.DefaultLocationID, u.ID)
	if err != nil {
		return err
	}
	err = u.updateUserDevices(txCtx, tx)
	if err != nil {
		return err
	}

	return tx.CommitTx(txCtx)
}

//go:embed user_device.update_by_mac.sql
var updateUserDeviceByMACSQL string

func (u *User) updateUserDevices(ctx context.Context, tx *database.Tx) error {
	for _, device := range u.Devices {
		err := tx.ExecTx(ctx, updateUserDeviceByMACSQL, device.Name, device.MAC)
		if err != nil {
			return err
		}
	}
	return nil
}

//go:embed user_device.insert.sql
var insertUserDeviceSQL string

// AddDevice creates and stores a new device for the given user.
func (u *User) AddDevice(ctx context.Context, driver *database.Driver, mac net.HardwareAddr, name string) (*UserDevice, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	device := &UserDevice{
		MAC:    mac.String(),
		Name:   name,
		UserID: u.ID,
	}
	err = tx.ExecTx(txCtx, insertUserDeviceSQL, device.MAC, device.Name, device.UserID)
	if err != nil {
		return nil, err
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}

	u.Devices = append(u.Devices, device)
	return device, nil
}

//go:embed user_device.delete_by_mac.sql
var deleteUserDeviceByMACSQL string

// RemoveDevice removes and deletes the given device for the given user.
func (u *User) RemoveDevice(ctx context.Context, driver *database.Driver, mac net.HardwareAddr) error {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	macString := mac.String()
	err = tx.ExecTx(txCtx, deleteUserDeviceByMACSQL, macString)
	if err != nil {
		return err
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return err
	}

	n := 0
	for _, device := range u.Devices {
		if device.MAC != macString {
			u.Devices[n] = device
			n++
		}
	}
	u.Devices = u.Devices[:n]
	return nil
}

//go:embed user.select.sql
var selectUserByIDSQL string

// SelectUsers load all stored users.
func SelectUserByID(ctx context.Context, driver *database.Driver, userID string) (*User, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	userRows, err := tx.QueryTx(txCtx, selectUserByIDSQL, userID)
	if err != nil {
		return nil, err
	}
	defer userRows.Close()
	var user *User
	if userRows.Next() {
		user = &User{}
		err = database.Scan(userRows, user)
		if err != nil {
			return nil, err
		}
		userDevices, err := user.selectUserDevices(txCtx, tx)
		if err != nil {
			return nil, err
		}
		user.Devices = userDevices
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

//go:embed user.select.sql
var selectUserSQL string

// SelectUsers load all stored users.
func SelectUsers(ctx context.Context, driver *database.Driver) ([]*User, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	users := make([]*User, 0)
	userRows, err := tx.QueryTx(txCtx, selectUserSQL)
	if err != nil {
		return nil, err
	}
	defer userRows.Close()
	for userRows.Next() {
		user := &User{}
		err = database.Scan(userRows, user)
		if err != nil {
			return nil, err
		}
		userDevices, err := user.selectUserDevices(txCtx, tx)
		if err != nil {
			return nil, err
		}
		user.Devices = userDevices
		users = append(users, user)
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

//go:embed user_device.select_by_user_id.sql
var selectUserDeviceByUserIDSQL string

func (u *User) selectUserDevices(ctx context.Context, tx *database.Tx) ([]*UserDevice, error) {
	userDeviceRows, err := tx.QueryTx(ctx, selectUserDeviceByUserIDSQL, u.ID)
	if err != nil {
		return nil, err
	}
	defer userDeviceRows.Close()
	userDevices := make([]*UserDevice, 0)
	for userDeviceRows.Next() {
		userDevice := &UserDevice{
			UserID: u.ID,
		}
		err = database.Scan(userDeviceRows, userDevice)
		if err != nil {
			return nil, err
		}
		userDevices = append(userDevices, userDevice)
	}
	return userDevices, nil
}

//go:embed user_device.delete_by_user_id.sql
var deleteUserDevicesByUserIDSQL string

//go:embed user.delete_by_id.sql
var deleteUserByIDSQL string

// Delete deletes the given user and any associated user device
func (u *User) Delete(ctx context.Context, driver *database.Driver) error {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	err = tx.ExecTx(txCtx, deleteUserDevicesByUserIDSQL, u.ID)
	if err != nil {
		return err
	}
	err = tx.ExecTx(txCtx, deleteUserByIDSQL, u.ID)
	if err != nil {
		return err
	}

	return tx.CommitTx(txCtx)
}
