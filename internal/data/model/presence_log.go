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
	"net/netip"

	"github.com/tdrn-org/go-database"
)

type PresenceLog struct {
	UserID     string `db:"user_id"`
	LocationID string `db:"location_id"`
	Active     bool   `db:"active"`
	IP         string `db:"ip"`
	Name       string `db:"name"`
	Timestamp  int64  `db:"timestamp"`
}

//go:embed presence_log.delete_by_pk.sql
var deletePresenceLogByPKSQL string

//go:embed presence_log.insert.sql
var insertPresenceLogSQL string

func LogPresence(ctx context.Context, driver *database.Driver, userID, locationID string, active bool, ip netip.Addr, name string) (*PresenceLog, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	err = tx.ExecTx(txCtx, deletePresenceLogByPKSQL, userID, locationID, active)
	if err != nil {
		return nil, err
	}
	log := &PresenceLog{
		UserID:     userID,
		LocationID: locationID,
		Active:     active,
		IP:         ip.StringExpanded(),
		Name:       name,
		Timestamp:  database.Now(),
	}
	err = tx.ExecTx(txCtx, insertPresenceLogSQL, log.UserID, log.LocationID, log.Active, log.IP, log.Name, log.Timestamp)
	if err != nil {
		return nil, err
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}

	return log, nil
}

//go:embed presence_log.select_by_user_id.sql
var selectPresenceLogByUserIDSQL string

func SelectPresenceLogByUserID(ctx context.Context, driver *database.Driver, userID string) ([]*PresenceLog, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	rows, err := tx.QueryTx(txCtx, selectPresenceLogByUserIDSQL, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	logs := make([]*PresenceLog, 0)
	for rows.Next() {
		log := &PresenceLog{
			UserID: userID,
		}
		err = database.Scan(rows, log)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

//go:embed presence_log.select_by_location_id.sql
var selectPresenceLogByLocationIDSQL string

func SelectPresenceLogByLocationID(ctx context.Context, driver *database.Driver, locationID string) ([]*PresenceLog, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	rows, err := tx.QueryTx(txCtx, selectPresenceLogByLocationIDSQL, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	logs := make([]*PresenceLog, 0)
	for rows.Next() {
		log := &PresenceLog{
			LocationID: locationID,
		}
		err = database.Scan(rows, log)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}

	return logs, nil
}
