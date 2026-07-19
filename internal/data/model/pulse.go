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

	"github.com/tdrn-org/go-database"
	"github.com/tdrn-org/totem/internal/domain/pulse"
)

type Pulse struct {
	ID          string `db:"id"`
	Agent       string `db:"agent"`
	Message     string `db:"message"`
	ScheduledAt int64  `db:"scheduled_at"`
	CreatedAt   int64  `db:"created_at"`
}

//go:embed pulse.insert.sql
var insertPulseSQL string

// InsertPulse inserts a new pulse.
func InsertPulse(ctx context.Context, driver *database.Driver, pulse *pulse.Pulse) (*Pulse, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	storePulse := &Pulse{
		ID:          database.NewID(),
		Agent:       pulse.Agent,
		Message:     pulse.Message,
		ScheduledAt: database.Time2DB(pulse.ScheduledAt),
		CreatedAt:   database.Time2DB(tx.Now()),
	}
	err = tx.ExecTx(txCtx, insertPulseSQL, storePulse.ID, storePulse.Agent, storePulse.Message, storePulse.ScheduledAt, storePulse.CreatedAt)
	if err != nil {
		return nil, err
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}

	pulse.ID, pulse.CreatedAt = storePulse.ID, database.DB2Time(storePulse.CreatedAt)
	return storePulse, nil
}

//go:embed pulse.select_by_agent.sql
var selectPulseByAgentSQL string

func SelectPulsesByAgent(ctx context.Context, driver *database.Driver, agent string) ([]*Pulse, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	rows, err := tx.QueryTx(txCtx, selectPulseByAgentSQL, agent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pulses := make([]*Pulse, 0)
	for rows.Next() {
		pulse := &Pulse{
			Agent: agent,
		}
		err = database.Scan(rows, pulse)
		if err != nil {
			return nil, err
		}
		pulses = append(pulses, pulse)
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}

	return pulses, nil
}

//go:embed pulse.select.sql
var selectPulseSQL string

func SelectPulses(ctx context.Context, driver *database.Driver) ([]*Pulse, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	rows, err := tx.QueryTx(txCtx, selectPulseSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pulses := make([]*Pulse, 0)
	for rows.Next() {
		pulse := &Pulse{}
		err = database.Scan(rows, pulse)
		if err != nil {
			return nil, err
		}
		pulses = append(pulses, pulse)
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}

	return pulses, nil
}

//go:embed pulse.delete_by_id.sql
var deletePulseByIDSQL string

func DeletePulsesByID(ctx context.Context, driver *database.Driver, id string) error {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	err = tx.ExecTx(txCtx, deletePulseByIDSQL, id)
	if err != nil {
		return err
	}

	return tx.CommitTx(txCtx)
}

//go:embed pulse.delete_by_agent.sql
var deletePulseByAgentSQL string

func DeletePulsesByAgent(ctx context.Context, driver *database.Driver, agent string) error {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	err = tx.ExecTx(txCtx, deletePulseByAgentSQL, agent)
	if err != nil {
		return err
	}

	return tx.CommitTx(txCtx)
}
