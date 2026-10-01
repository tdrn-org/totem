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

	"github.com/tdrn-org/go-database"
	"github.com/tdrn-org/totem/internal/data/model"
	"github.com/tdrn-org/totem/internal/domain/pulse"
)

func (s *Store) SavePulse(ctx context.Context, pulse *pulse.Pulse) error {
	txCtx, tx, err := s.driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	_, err = model.InsertPulse(txCtx, s.driver, pulse)
	if err != nil {
		return err
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ListAgentPulses(ctx context.Context, agent string) ([]*pulse.Pulse, error) {
	storePulses, err := model.SelectPulsesByAgent(ctx, s.driver, agent)
	if err != nil {
		return nil, err
	}
	pulses := make([]*pulse.Pulse, 0, len(storePulses))
	for _, storPulse := range storePulses {
		pulse := s.model2Pulse(storPulse)
		pulses = append(pulses, pulse)
	}
	return pulses, nil
}

func (s *Store) model2Pulse(storePulse *model.Pulse) *pulse.Pulse {
	if storePulse == nil {
		return nil
	}
	pulse := &pulse.Pulse{
		ID:          storePulse.ID,
		Agent:       storePulse.Agent,
		Message:     storePulse.Message,
		ScheduledAt: database.DB2Time(storePulse.ScheduledAt),
		CreatedAt:   database.DB2Time(storePulse.CreatedAt),
	}
	return pulse
}

func (s *Store) ListPulses(ctx context.Context) ([]*pulse.Pulse, error) {
	storePulses, err := model.SelectPulses(ctx, s.driver)
	if err != nil {
		return nil, err
	}
	pulses := make([]*pulse.Pulse, 0, len(storePulses))
	for _, storPulse := range storePulses {
		pulse := s.model2Pulse(storPulse)
		pulses = append(pulses, pulse)
	}
	return pulses, nil
}

func (s *Store) DeletePulse(ctx context.Context, id string) error {
	return model.DeletePulsesByID(ctx, s.driver, id)
}
