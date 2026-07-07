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
	"bytes"
	"context"
	"log/slog"

	"github.com/tdrn-org/totem/internal/adapters/integration"
	"github.com/tdrn-org/totem/internal/domain"
	"github.com/tdrn-org/totem/internal/domain/lan"
	"github.com/tdrn-org/totem/internal/domain/presence"
	"github.com/tdrn-org/totem/internal/domain/smarthome"
)

type PresenceTracker struct {
	host   Host
	logger *slog.Logger
}

func NewPresenceTracker(host Host) *PresenceTracker {
	return &PresenceTracker{
		host:   host,
		logger: slog.With("application", "presenceTracker"),
	}
}

func (t *PresenceTracker) Track(ctx context.Context) {
	t.logger.Info("updating tracking logs...")
	for _, service := range t.host.Services() {
		for _, nature := range service.Natures() {
			switch nature {
			case integration.NatureLan:
				t.trackLanService(ctx, service.ID(), service.(lan.LanService))
			}
		}
	}
}

func (t *PresenceTracker) trackLanService(ctx context.Context, serviceID string, service lan.LanService) {
	t.logger.Debug("probing Lan devices", slog.String("serviceID", serviceID))
	lanDevices, err := service.ListDevices(ctx)
	if err != nil {
		t.logger.Warn("failed to list Lan devices", slog.String("service", serviceID), slog.Any("err", err))
		return
	}
	users, err := t.host.DataStore().ListUsers(ctx)
	if err != nil {
		t.logger.Warn("failed to list users", slog.Any("err", err))
		return
	}
	for _, user := range users {
		for _, userDevice := range user.Devices {
			lanDevice := t.matchUserDevice(userDevice, lanDevices)
			if lanDevice != nil {
				t.logger.Debug("logging user presence", slog.String("user", user.Name), slog.Bool("active", lanDevice.Active))
				_, err = t.host.DataStore().LogUserPresence(ctx, user, lanDevice.Location, lanDevice.Active, lanDevice.IP, lanDevice.Name)
				if err != nil {
					t.logger.Warn("failed to log user presence", slog.Any("err", err))
				}
			}
		}
	}
}

func (t *PresenceTracker) matchUserDevice(userDevice *domain.UserDevice, lanDevices []*lan.LanDevice) *lan.LanDevice {
	for _, lanDevice := range lanDevices {
		if bytes.Equal(userDevice.MAC, lanDevice.MAC) {
			return lanDevice
		}
	}
	return nil
}

func (t *PresenceTracker) CheckUser(ctx context.Context, user *domain.User) (*presence.UserPresence, error) {
	logs, err := t.host.DataStore().ListUserPresenceLogs(ctx, user, false)
	if err != nil {
		return nil, err
	}
	result := &presence.UserPresence{}
	if len(logs) == 0 {
		return result, nil
	}

	// Overall newest entry for LastSeen fallback
	result.LastSeenLocation = logs[0].Location
	if !logs[0].Timestamp.IsZero() {
		t := logs[0].Timestamp
		result.LastSeen = &t
	}

	// Per location: first occurrence is the newest for that location
	seenLocations := make(map[string]bool)
	for _, log := range logs {
		locID := log.Location.ID
		if seenLocations[locID] {
			continue // Already saw this location's newest entry
		}
		seenLocations[locID] = true
		if log.Active {
			result.Location = log.Location
			break
		}
	}

	return result, nil
}

func (t *PresenceTracker) CheckLocation(ctx context.Context, location *domain.Location) (*presence.PresenceResult, error) {
	logs, err := t.host.DataStore().ListLocationPresenceLogs(ctx, location)
	if err != nil {
		return nil, err
	}
	users := make([]*domain.User, 0, len(logs))
	for _, log := range logs {
		if !log.Active {
			continue
		}
		users = append(users, log.User)
	}
	var motionSensors []*smarthome.MotionSensor
	for _, service := range t.host.Services() {
		for _, nature := range service.Natures() {
			if nature == integration.NatureSmarthome {
				smarthomeService := service.(smarthome.SmarthomeService)
				if smarthomeService.Location().Equal(location) {
					allSensors, err := smarthomeService.ListMotionSensors(ctx)
					if err != nil {
						return nil, err
					}
					for _, sensor := range allSensors {
						if sensor.Motion {
							motionSensors = append(motionSensors, sensor)
						}
					}
				}
			}
		}
	}
	result := &presence.PresenceResult{
		Users:         users,
		MotionSensors: motionSensors,
	}
	return result, nil
}
