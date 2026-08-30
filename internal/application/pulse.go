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

package application

import (
	"context"
	"log/slog"
	"time"
)

type PulseScheduler struct {
	host   Host
	logger *slog.Logger
}

func NewPulseScheduler(host Host) *PulseScheduler {
	return &PulseScheduler{
		host:   host,
		logger: slog.With("application", "pulseScheduler"),
	}
}

func (t *PulseScheduler) RunPulses(ctx context.Context, last, now time.Time) error {
	t.logger.Info("running pulses...")
	pulses, err := t.host.DataStore().ListPulses(ctx)
	if err != nil {
		t.logger.Warn("failed to list Pulses", slog.Any("err", err))
		return err
	}
	for _, pulse := range pulses {
		if pulse.ScheduledAt.Before(now) {
			err := t.host.DataStore().DeletePulse(ctx, pulse.ID)
			if err != nil {
				t.logger.Warn("failed to delete Pulse", slog.Any("err", err))
				continue
			}
			t.host.NotificationService().PostPulse(ctx, pulse)
		}
	}
	return nil
}
