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
	driver            *database.Driver
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

func NewUser(driver *database.Driver, name, defaultLocationID string) *User {
	return &User{
		driver:            driver,
		ID:                database.NewID(),
		Name:              name,
		DefaultLocationID: defaultLocationID,
		Devices:           []*UserDevice{},
	}
}

//go:embed user.insert.sql
var insertUserSQL string

func (u *User) Insert(ctx context.Context) error {
	txCtx, tx, err := u.driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	args0 := []any{
		u.ID,
		u.Name,
		u.DefaultLocationID,
	}
	err = tx.ExecTx(txCtx, insertUserSQL, args0...)
	if err != nil {
		return err
	}

	return tx.CommitTx(txCtx)
}

//go:embed user_device.insert.sql
var insertUserDeviceSQL string

func (u *User) AddDevice(ctx context.Context, mac net.HardwareAddr, name string) (*UserDevice, error) {
	txCtx, tx, err := u.driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	device := &UserDevice{
		MAC:    mac.String(),
		Name:   name,
		UserID: u.ID,
	}
	args0 := []any{
		device.MAC,
		device.Name,
		device.UserID,
	}
	err = tx.ExecTx(txCtx, insertUserDeviceSQL, args0...)
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

func SelectUsers(ctx context.Context, driver *database.Driver) ([]*User, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	users := make([]*User, 0)

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}

	return users, nil
}
